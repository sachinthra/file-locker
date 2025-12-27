package main

import (
	"context"
	"fmt"
	"log"
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
	grpcService "github.com/sachinthra/file-locker/backend/internal/grpc"
	"github.com/sachinthra/file-locker/backend/internal/storage"
	"github.com/sachinthra/file-locker/backend/internal/worker"
	pb "github.com/sachinthra/file-locker/backend/pkg/proto"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting File Locker Backend...")
	log.Printf("HTTP Server will run on port %d", cfg.Server.Port)
	log.Printf("gRPC Server will run on port %d", cfg.Server.GRPCPort)

	// Initialize storage services
	log.Println("Initializing storage services...")

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
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}
	log.Println("✓ MinIO connected successfully")

	// Initialize Redis
	redisCache, err := storage.NewRedisCache(
		cfg.Storage.Redis.Addr,
		cfg.Storage.Redis.Password,
		cfg.Storage.Redis.DB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	log.Println("✓ Redis connected successfully")

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		cfg.Security.JWTSecret,
		cfg.Security.SessionTimeout,
	)
	log.Println("✓ JWT service initialized")

	// Initialize auth middleware
	authMiddleware := auth.NewAuthMiddleware(jwtService, redisCache)

	// Initialize API handlers
	authHandler := api.NewAuthHandler(jwtService, redisCache)
	uploadHandler := api.NewUploadHandler(minioStorage, redisCache)
	downloadHandler := api.NewDownloadHandler(minioStorage, redisCache)
	streamHandler := api.NewStreamHandler(minioStorage, redisCache)
	filesHandler := api.NewFilesHandler(redisCache)

	log.Println("✓ API handlers initialized")

	// Setup HTTP Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:9010", "null"},
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:9010"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders: []string{"Content-Length", "Content-Range"},
		// AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "Range"},
		// ExposedHeaders:   []string{"Content-Length", "Content-Range", "Accept-Ranges", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// You also need to serve the static YAML file itself
	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.yaml")
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:9010/docs/openapi.yaml"), // pointing to your YAML
	))

	log.Println("✓ Swagger documentation endpoint configured at /swagger/index.html")

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
			r.Delete("/files", filesHandler.HandleDeleteFile)
			r.Get("/download/{id}", downloadHandler.HandleDownload)
			r.Get("/stream/{id}", streamHandler.HandleStream)

			// Auth operations
			r.Post("/auth/logout", authHandler.HandleLogout)
		})
	})

	log.Println("✓ HTTP routes configured")

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
	fileServiceServer := grpcService.NewFileServiceServer(redisCache)
	pb.RegisterFileServiceServer(grpcServer, fileServiceServer)
	log.Println("✓ gRPC server initialized")

	// Start cleanup worker if enabled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Features.AutoDelete.Enabled {
		cleanupInterval := time.Duration(cfg.Features.AutoDelete.CheckInterval) * time.Minute
		cleanupWorker := worker.NewCleanupWorker(minioStorage, redisCache, cleanupInterval)
		go cleanupWorker.Start(ctx)
		log.Printf("✓ Cleanup worker started (interval: %v)", cleanupInterval)
	}

	// Start gRPC server in a goroutine
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	go func() {
		log.Printf("🚀 gRPC server listening on :%d", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
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
		log.Printf("🚀 HTTP server listening on :%d", cfg.Server.Port)
		log.Println("=====================================")
		log.Println("File Locker Backend is ready!")
		log.Println("=====================================")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Cancel background workers
	cancel()

	// Gracefully shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()

	log.Println("Servers stopped gracefully")
}
