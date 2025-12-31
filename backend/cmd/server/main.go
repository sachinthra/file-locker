package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sachinthra/file-locker/backend/internal/api"
	"github.com/sachinthra/file-locker/backend/internal/auth"
	"github.com/sachinthra/file-locker/backend/internal/config"
	"github.com/sachinthra/file-locker/backend/internal/db"
	grpcService "github.com/sachinthra/file-locker/backend/internal/grpc"
	"github.com/sachinthra/file-locker/backend/internal/logger"
	"github.com/sachinthra/file-locker/backend/internal/storage"
	"github.com/sachinthra/file-locker/backend/internal/worker"
	pb "github.com/sachinthra/file-locker/backend/pkg/proto"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration (with strict validation)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	appLogger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("❌ Failed to initialize logger: %v", err)
	}

	appLogger.Info("Starting File Locker Backend",
		slog.Int("http_port", cfg.Server.Port),
		slog.Int("grpc_port", cfg.Server.GRPCPort),
		slog.String("log_level", cfg.Logging.Level),
	)

	// Run database migrations
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Storage.Database.User,
		cfg.Storage.Database.Password,
		cfg.Storage.Database.Host,
		cfg.Storage.Database.Port,
		cfg.Storage.Database.DBName,
	)

	appLogger.Info("Running database migrations")
	if err := db.RunMigrations(dbURL, appLogger); err != nil {
		appLogger.Error("Failed to run database migrations", slog.String("error", err.Error()))
		log.Fatalf("❌ Database migration failed: %v\nPlease check your database configuration and try again.", err)
	}
	appLogger.Info("✅ Database migrations completed successfully")

	// Create default admin user
	if err := db.CreateDefaultAdmin(
		dbURL,
		cfg.Security.DefaultAdmin.Username,
		cfg.Security.DefaultAdmin.Email,
		cfg.Security.DefaultAdmin.Password,
		appLogger,
	); err != nil {
		appLogger.Error("Failed to create default admin", slog.String("error", err.Error()))
		log.Fatalf("❌ Failed to create default admin: %v", err)
	}

	// Initialize storage services
	appLogger.Info("Initializing storage services")

	// Initialize PostgreSQL
	pgStore, err := storage.NewPostgresStore(
		cfg.Storage.Database.Host,
		fmt.Sprintf("%d", cfg.Storage.Database.Port),
		cfg.Storage.Database.User,
		cfg.Storage.Database.Password,
		cfg.Storage.Database.DBName,
	)
	if err != nil {
		appLogger.Error("Failed to initialize PostgreSQL", slog.String("error", err.Error()))
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}
	appLogger.Info("PostgreSQL connected successfully",
		slog.String("host", cfg.Storage.Database.Host),
		slog.String("database", cfg.Storage.Database.DBName),
	)
	defer pgStore.Close()

	// Initialize MinIO
	minioStorage, err := storage.NewMinIOStorage(
		cfg.Storage.MinIO.Endpoint,
		cfg.Storage.MinIO.AccessKey,
		cfg.Storage.MinIO.SecretKey,
		cfg.Storage.MinIO.Bucket,
		cfg.Storage.MinIO.UseSSL,
		cfg.Storage.MinIO.Region,
	)
	if err != nil {
		appLogger.Error("Failed to initialize MinIO", slog.String("error", err.Error()))
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}
	appLogger.Info("MinIO connected successfully",
		slog.String("endpoint", cfg.Storage.MinIO.Endpoint),
		slog.String("bucket", cfg.Storage.MinIO.Bucket),
	)

	// Initialize Redis
	redisCache, err := storage.NewRedisCache(
		cfg.Storage.Redis.Addr,
		cfg.Storage.Redis.Password,
		cfg.Storage.Redis.DB,
	)
	if err != nil {
		appLogger.Error("Failed to initialize Redis", slog.String("error", err.Error()))
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	appLogger.Info("Redis connected successfully", slog.String("addr", cfg.Storage.Redis.Addr))

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		cfg.Security.JWTSecret,
		cfg.Security.SessionTimeout,
	)
	appLogger.Info("JWT service initialized")

	// Initialize auth middleware
	authMiddleware := auth.NewAuthMiddleware(jwtService, redisCache, pgStore)

	// Initialize API handlers
	authHandler := api.NewAuthHandler(jwtService, redisCache, pgStore)
	userHandler := api.NewUserHandler(pgStore)
	tokensHandler := api.NewTokensHandler(pgStore)
	uploadHandler := api.NewUploadHandler(minioStorage, redisCache, pgStore)
	downloadHandler := api.NewDownloadHandler(minioStorage, redisCache, pgStore)
	streamHandler := api.NewStreamHandler(minioStorage, redisCache, pgStore)
	filesHandler := api.NewFilesHandler(redisCache, minioStorage, pgStore)
	exportHandler := api.NewExportHandler(minioStorage, pgStore)
	adminHandler := api.NewAdminHandler(pgStore, minioStorage, redisCache)

	appLogger.Info("API handlers initialized")

	// Setup HTTP Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware (frontend accessed through nginx on port 80)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost", "http://localhost:80", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "X-Real-IP", "X-Forwarded-For"},
		ExposedHeaders:   []string{"Content-Length", "Content-Range"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint (supports both GET and HEAD)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	r.Head("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// You also need to serve the static YAML file itself
	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:9010/docs/openapi.yaml"), // pointing to your YAML
	))

	appLogger.Info("Swagger documentation configured", slog.String("endpoint", "/swagger/index.html"))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no authentication required)
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", authHandler.HandleLogin)
			r.Post("/auth/register", authHandler.HandleRegister)
		})

		// Protected routes (authentication required)
		r.Group(func(r chi.Router) {
			// Apply auth middleware
			r.Use(authMiddleware.RequireAuth)

			// Apply rate limiting if enabled
			if cfg.Security.RateLimit.Enabled {
				r.Use(authMiddleware.RateLimitMiddleware(
					cfg.Security.RateLimit.RequestsPerMinute,
					1*time.Minute,
				))
			}

			// File operations
			r.Post("/upload", uploadHandler.HandleUpload)
			r.Get("/files", filesHandler.HandleListFiles)
			r.Get("/files/search", filesHandler.HandleSearchFiles)
			r.Get("/files/export", exportHandler.HandleExportAll)
			r.Delete("/files", filesHandler.HandleDeleteFile)
			r.Patch("/files/{fileID}", filesHandler.HandleUpdateFile)
			r.Get("/download/{id}", downloadHandler.HandleDownload)
			r.Get("/stream/{id}", streamHandler.HandleStream)

			// User operations
			r.Patch("/user/password", userHandler.HandleChangePassword)

			// Auth operations
			r.Post("/auth/logout", authHandler.HandleLogout)
			r.Get("/auth/me", authHandler.HandleGetMe)

			// Personal Access Tokens (PATs)
			r.Post("/auth/tokens", tokensHandler.HandleCreateToken)
			r.Get("/auth/tokens", tokensHandler.HandleListTokens)
			r.Delete("/auth/tokens/{id}", tokensHandler.HandleRevokeToken)

			// Announcements (user operations)
			r.Get("/announcements", adminHandler.HandleGetAnnouncements)
			r.Post("/announcements/{id}/dismiss", adminHandler.HandleDismissAnnouncement)
		})

		// Admin routes (authentication + admin role required)
		r.Group(func(r chi.Router) {
			// Apply auth middleware
			r.Use(authMiddleware.RequireAuth)
			// Apply admin-only middleware
			r.Use(authMiddleware.RequireAdmin)

			// System statistics
			r.Get("/admin/stats", adminHandler.HandleGetStats)

			// User management
			r.Get("/admin/users", adminHandler.HandleGetUsers)
			r.Get("/admin/users/pending", adminHandler.HandleGetPendingUsers)
			r.Post("/admin/users/{id}/approve", adminHandler.HandleApproveUser)
			r.Post("/admin/users/{id}/reject", adminHandler.HandleRejectUser)
			r.Delete("/admin/users/{id}", adminHandler.HandleDeleteUser)
			r.Patch("/admin/users/{id}/status", adminHandler.HandleUpdateUserStatus)
			r.Patch("/admin/users/{id}/role", adminHandler.HandleUpdateUserRole)
			r.Post("/admin/users/{id}/reset-password", adminHandler.HandleResetUserPassword)
			r.Post("/admin/users/{id}/logout", adminHandler.HandleForceLogoutUser)

			// Settings management
			r.Get("/admin/settings", adminHandler.HandleGetSettings)
			r.Patch("/admin/settings", adminHandler.HandleUpdateSetting)

			// Announcements management
			r.Get("/admin/announcements", adminHandler.HandleGetAnnouncements)
			r.Post("/admin/announcements", adminHandler.HandleCreateAnnouncement)
			r.Delete("/admin/announcements/{id}", adminHandler.HandleDeleteAnnouncement)

			// Global file management
			r.Get("/admin/files", adminHandler.HandleGetAllFiles)
			r.Delete("/admin/files/{id}", adminHandler.HandleDeleteAnyFile)

			// Storage cleanup
			r.Get("/admin/storage/analyze", adminHandler.HandleAnalyzeStorage)
			r.Post("/admin/storage/cleanup", adminHandler.HandleCleanupStorage)

			// Audit logs
			r.Get("/admin/logs", adminHandler.HandleGetAuditLogs)
		})
	})

	appLogger.Info("HTTP routes configured")

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	fileServiceServer := grpcService.NewFileServiceServer(pgStore)
	pb.RegisterFileServiceServer(grpcServer, fileServiceServer)
	appLogger.Info("gRPC server initialized")

	// Start cleanup worker if enabled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Features.AutoDelete.Enabled {
		cleanupInterval := time.Duration(cfg.Features.AutoDelete.CheckInterval) * time.Minute
		cleanupWorker := worker.NewCleanupWorker(minioStorage, pgStore, cleanupInterval)
		go cleanupWorker.Start(ctx)
		appLogger.Info("Cleanup worker started", slog.Duration("interval", cleanupInterval))
	}

	// Start gRPC server in a goroutine
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	go func() {
		appLogger.Info("🚀 gRPC server listening", slog.Int("port", cfg.Server.GRPCPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			appLogger.Error("gRPC server failed", slog.String("error", err.Error()))
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Start HTTP server
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        r,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start HTTP server in a goroutine
	go func() {
		appLogger.Info("🚀 HTTP server listening", slog.Int("port", cfg.Server.Port))
		appLogger.Info("File Locker Backend is ready!")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error("HTTP server failed", slog.String("error", err.Error()))
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down servers...")

	// Cancel background workers
	cancel()

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("HTTP server forced to shutdown", slog.String("error", err.Error()))
	}

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()

	appLogger.Info("Servers stopped gracefully")
}
