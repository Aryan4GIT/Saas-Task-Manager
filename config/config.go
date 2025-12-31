package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Gemini   GeminiConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port      string
	Env       string
	UploadDir string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type GeminiConfig struct {
	APIKey         string
	Model          string
	EmbeddingModel string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	// Try to load .env file from multiple locations
	// First try current directory, then walk up to find the project root
	envPaths := []string{
		".env",
		"../../.env", // When running from cmd/server
		"../.env",    // When running from cmd
	}

	for _, path := range envPaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				_ = godotenv.Load(absPath)
				break
			}
		}
	}

	env := getEnv("ENV", "development")

	// Parse JWT expiry durations
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXPIRY: %w", err)
	}

	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXPIRY: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port:      getEnv("PORT", "8080"),
			Env:       env,
			UploadDir: getEnv("UPLOAD_DIR", "uploads"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "saas_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			AccessSecret:  "",
			RefreshSecret: "",
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Gemini: GeminiConfig{
			APIKey:         getEnv("GEMINI_API_KEY", ""),
			Model:          getEnv("GEMINI_MODEL", "gemini-1.5-pro"),
			EmbeddingModel: getEnv("GEMINI_EMBEDDING_MODEL", "text-embedding-004"),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseList(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		},
	}

	// JWT secrets: required in production; auto-default in development to reduce setup friction.
	config.JWT.AccessSecret = getEnv("JWT_ACCESS_SECRET", "")
	config.JWT.RefreshSecret = getEnv("JWT_REFRESH_SECRET", "")
	if env != "production" {
		if config.JWT.AccessSecret == "" {
			config.JWT.AccessSecret = "dev-access-secret-change-me"
		}
		if config.JWT.RefreshSecret == "" {
			config.JWT.RefreshSecret = "dev-refresh-secret-change-me"
		}
	}

	// Validate required fields
	if config.Server.Env == "production" {
		if config.JWT.AccessSecret == "" {
			return nil, fmt.Errorf("JWT_ACCESS_SECRET is required")
		}
		if config.JWT.RefreshSecret == "" {
			return nil, fmt.Errorf("JWT_REFRESH_SECRET is required")
		}
	}

	return config, nil
}

func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseList(s string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, char := range s {
		if char == ',' {
			if current != "" {
				result = append(result, strings.TrimSpace(current))
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, strings.TrimSpace(current))
	}
	return result
}
