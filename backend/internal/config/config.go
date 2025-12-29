package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	Security SecurityConfig `mapstructure:"security" validate:"required"`
	Storage  StorageConfig  `mapstructure:"storage" validate:"required"`
	Features FeaturesConfig `mapstructure:"features" validate:"required"`
	Logging  LoggingConfig  `mapstructure:"logging" validate:"required"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	GRPCPort       int           `mapstructure:"grpc_port" validate:"required,min=1,max=65535"`
	Host           string        `mapstructure:"host" validate:"required"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" validate:"required"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes" validate:"required,min=1"`
}

type SecurityConfig struct {
	JWTSecret      string          `mapstructure:"jwt_secret" validate:"required,min=16"`
	SessionTimeout int             `mapstructure:"session_timeout" validate:"required,min=60"`
	DefaultAdmin   DefaultAdmin    `mapstructure:"default_admin" validate:"required"`
	TLS            TLSConfig       `mapstructure:"tls" validate:"required"`
	RateLimit      RateLimitConfig `mapstructure:"rate_limiting" validate:"required"`
}

type DefaultAdmin struct {
	Username string `mapstructure:"username" validate:"required,min=3"`
	Email    string `mapstructure:"email" validate:"required,email"`
	Password string `mapstructure:"password" validate:"required,min=8"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute" validate:"min=0"`
	Burst             int  `mapstructure:"burst" validate:"min=0"`
}

type StorageConfig struct {
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	MinIO    MinIOConfig    `mapstructure:"minio" validate:"required"`
	Redis    RedisConfig    `mapstructure:"redis" validate:"required"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host" validate:"required"`
	Port            int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	User            string `mapstructure:"user" validate:"required"`
	Password        string `mapstructure:"password" validate:"required"`
	DBName          string `mapstructure:"dbname" validate:"required"`
	SSLMode         string `mapstructure:"sslmode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" validate:"required,min=1"`
}

type MinIOConfig struct {
	Endpoint    string `mapstructure:"endpoint" validate:"required"`
	PortAPI     int    `mapstructure:"port_api" validate:"required,min=1,max=65535"`     // For Docker Port Mapping
	PortConsole int    `mapstructure:"port_console" validate:"required,min=1,max=65535"` // For Docker Port Mapping
	AccessKey   string `mapstructure:"access_key" validate:"required"`
	SecretKey   string `mapstructure:"secret_key" validate:"required"`
	Bucket      string `mapstructure:"bucket" validate:"required"`
	UseSSL      bool   `mapstructure:"use_ssl"`
	Region      string `mapstructure:"region" validate:"required"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"` // For Docker Port Mapping
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
}

type FeaturesConfig struct {
	AutoDelete     AutoDeleteConfig     `mapstructure:"auto_delete" validate:"required"`
	VideoStreaming VideoStreamingConfig `mapstructure:"video_streaming" validate:"required"`
	BatchUploads   BatchUploadsConfig   `mapstructure:"batch_uploads" validate:"required"`
}

type AutoDeleteConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	CheckInterval int  `mapstructure:"check_interval" validate:"min=1"`
}

type VideoStreamingConfig struct {
	Enabled   bool `mapstructure:"enabled"`
	ChunkSize int  `mapstructure:"chunk_size" validate:"min=1"`
}

type BatchUploadsConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	MaxConcurrent int  `mapstructure:"max_concurrent" validate:"min=1"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Path       string `mapstructure:"path" validate:"required"`
	MaxSizeMB  int    `mapstructure:"max_size_mb" validate:"min=1"`
	MaxBackups int    `mapstructure:"max_backups" validate:"min=1"`
	MaxAgeDays int    `mapstructure:"max_age_days" validate:"min=1"`
}

// LoadConfig loads configuration from file and environment
func LoadConfig() (*Config, error) {
	viper.SetConfigType("yaml")

	// 1. Check for explicit path via Environment Variable (Used by Docker/Makefile)
	configPath := os.Getenv("CONFIG_PATH")

	if configPath != "" {
		viper.SetConfigFile(configPath)
		fmt.Printf("🔍 Loading configuration from CONFIG_PATH: %s\n", configPath)
	} else {
		viper.AddConfigPath(".") // Check current directory
		// Default paths for local development (go run main.go)
		viper.AddConfigPath("./configs") // Check ./configs
		// viper.AddConfigPath("../configs") // Check ../configs (if running from cmd/)

		viper.AddConfigPath("/etc/file-locker")
		viper.AddConfigPath("/usr/local/etc/file-locker")
		viper.AddConfigPath("/opt/file-locker/configs")
	}

	// 2. Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	fmt.Printf("✅ Configuration loaded from: %s\n", viper.ConfigFileUsed())

	// 3. Setup Environment Variable Overrides
	// This allows Docker to inject "minio:9000" instead of "localhost:9012"
	viper.SetEnvPrefix("FILELOCKER")

	// Crucial: Replace dots with underscores (storage.minio.endpoint -> STORAGE_MINIO_ENDPOINT)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 4. Strict Validation (Fail Fast)
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		// Format validation errors with detailed messages
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMessages []string
			for _, fieldErr := range validationErrors {
				errMessages = append(errMessages, fmt.Sprintf(
					"❌ Field '%s' failed validation: %s (value: '%v')",
					fieldErr.Namespace(),
					fieldErr.Tag(),
					fieldErr.Value(),
				))
			}
			return nil, fmt.Errorf("configuration validation failed:\n%s", strings.Join(errMessages, "\n"))
		}
		return nil, fmt.Errorf("validation error: %w", err)
	}

	fmt.Println("✅ Configuration validation passed")
	return &config, nil
}
