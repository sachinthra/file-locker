package config

import (
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
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
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
	// 1. Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add multiple possible config paths to handle different working directories
	viper.AddConfigPath(".")                // Config in current directory
	viper.AddConfigPath("./configs")        // Running from project root
	viper.AddConfigPath("../configs")       // Running from backend/
	viper.AddConfigPath("../../configs")    // Running from backend/internal/
	viper.AddConfigPath("../../../configs") // Running from backend/internal/config/

	// 2. Read the config file
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional, just log if not found
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 3. Also try to load .env file (optional) - try multiple paths
	envPaths := []string{".env", "../.env", "../../.env", "../../../.env"}
	for _, envPath := range envPaths {
		viper.SetConfigFile(envPath)
		viper.MergeInConfig() // Merge instead of replace, ignore errors
	}

	// 4. Read environment variables (override file values)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FILELOCKER")

	// 5. Set defaults
	viper.SetDefault("server.port", 9010)
	viper.SetDefault("server.grpc_port", 9011)
	viper.SetDefault("server.host", "0.0.0.0")

	// 6. Unmarshal into Config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
