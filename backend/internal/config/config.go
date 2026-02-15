package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	SSH      SSHConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MinConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	BcryptCost    int
}

type SSHConfig struct {
	MaxConcurrentSessions int
	SessionIdleTimeout    time.Duration
	SessionMaxDuration    time.Duration
	MaxBufferSize         int
}

type SecurityConfig struct {
	EncryptionKey        string
	AllowedOrigins       []string
	RateLimitPerMinute   int
	MaxLoginAttempts     int
	LoginBlockDuration   time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "webssh"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
			MaxConns: getIntEnv("DB_MAX_CONNS", 100),
			MinConns: getIntEnv("DB_MIN_CONNS", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
			JWTExpiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
			BcryptCost:    getIntEnv("BCRYPT_COST", 10),
		},
		SSH: SSHConfig{
			MaxConcurrentSessions: getIntEnv("SSH_MAX_CONCURRENT_SESSIONS", 1000),
			SessionIdleTimeout:    getDurationEnv("SSH_SESSION_IDLE_TIMEOUT", 30*time.Minute),
			SessionMaxDuration:    getDurationEnv("SSH_SESSION_MAX_DURATION", 4*time.Hour),
			MaxBufferSize:         getIntEnv("SSH_MAX_BUFFER_SIZE", 10*1024*1024), // 10MB
		},
		Security: SecurityConfig{
			EncryptionKey:      getEnv("ENCRYPTION_KEY", "change-me-32-chars-minimum!!!"),
			AllowedOrigins:     getSliceEnv("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
			RateLimitPerMinute: getIntEnv("RATE_LIMIT_PER_MINUTE", 100),
			MaxLoginAttempts:   getIntEnv("MAX_LOGIN_ATTEMPTS", 5),
			LoginBlockDuration: getDurationEnv("LOGIN_BLOCK_DURATION", 15*time.Minute),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Security.EncryptionKey) < 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be at least 32 characters")
	}
	if c.Auth.JWTSecret == "change-me-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Change in production!")
	}
	return nil
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return []string{value}
	}
	return defaultValue
}
