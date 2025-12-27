package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Security SecurityConfig `mapstructure:"security"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Features FeaturesConfig `mapstructure:"features"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port"`
	GRPCPort       int           `mapstructure:"grpc_port"`
	Host           string        `mapstructure:"host"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

type SecurityConfig struct {
	JWTSecret      string          `mapstructure:"jwt_secret"`
	SessionTimeout int             `mapstructure:"session_timeout"`
	TLS            TLSConfig       `mapstructure:"tls"`
	RateLimit      RateLimitConfig `mapstructure:"rate_limiting"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

type StorageConfig struct {
	MinIO MinIOConfig `mapstructure:"minio"`
	Redis RedisConfig `mapstructure:"redis"`
}

type MinIOConfig struct {
	Endpoint    string `mapstructure:"endpoint"`
	PortAPI     int    `mapstructure:"port_api"`     // For Docker Port Mapping
	PortConsole int    `mapstructure:"port_console"` // For Docker Port Mapping
	AccessKey   string `mapstructure:"access_key"`
	SecretKey   string `mapstructure:"secret_key"`
	Bucket      string `mapstructure:"bucket"`
	UseSSL      bool   `mapstructure:"use_ssl"`
	Region      string `mapstructure:"region"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Port     int    `mapstructure:"port"` // For Docker Port Mapping
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type FeaturesConfig struct {
	AutoDelete     AutoDeleteConfig     `mapstructure:"auto_delete"`
	VideoStreaming VideoStreamingConfig `mapstructure:"video_streaming"`
	BatchUploads   BatchUploadsConfig   `mapstructure:"batch_uploads"`
}

type AutoDeleteConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	CheckInterval int  `mapstructure:"check_interval"`
}

type VideoStreamingConfig struct {
	Enabled   bool `mapstructure:"enabled"`
	ChunkSize int  `mapstructure:"chunk_size"`
}

type BatchUploadsConfig struct {
	Enabled       bool `mapstructure:"enabled"`
	MaxConcurrent int  `mapstructure:"max_concurrent"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
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
		return nil, err
	}

	return &config, nil
}
