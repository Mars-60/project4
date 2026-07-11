package configs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func Load() error {

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to load .env file: %w", err)
	}

	App = AppConfig{
		Name:    getenv("APP_NAME", "TradePilot AI"),
		Version: getenv("APP_VERSION", "v0.1.0"),
		Env:     getenv("APP_ENV", "development"),

		Server: ServerConfig{
			Host:            getenv("SERVER_HOST", "localhost"),
			Port:            getenv("SERVER_PORT", "8080"),
			ReadTimeout:     getDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
			CORS: CORSConfig{
				AllowedOrigins: getCSV("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
				AllowedMethods: getCSV("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS"),
				AllowedHeaders: getCSV("CORS_ALLOWED_HEADERS", "Accept,Authorization,Content-Type,X-CSRF-Token,X-Request-ID"),
			},
		},

		Log: LogConfig{
			Level: getenv("LOG_LEVEL", "debug"),
		},

		DB: DatabaseConfig{
			Driver:          getenv("DB_DRIVER", "pgx"),
			DSN:             getenv("DATABASE_URL", ""),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		},

		Auth: AuthConfig{
			JWTSecret:          getenv("JWT_SECRET", "change-me-in-production"),
			AccessTokenTTL:     getDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL:    getDuration("JWT_REFRESH_TTL", 720*time.Hour),
			PasswordIterations: getInt("PASSWORD_HASH_ITERATIONS", 120000),
		},

		Groq: GroqConfig{
			BaseURL:    strings.TrimRight(getenv("GROQ_BASE_URL", "https://api.groq.com/openai/v1"), "/"),
			APIKey:     getenv("GROQ_API_KEY", ""),
			Model:      getenv("GROQ_MODEL", "llama-3.3-70b-versatile"),
			Timeout:    getDuration("GROQ_TIMEOUT", 30*time.Second),
			MaxRetries: getInt("GROQ_MAX_RETRIES", 3),
		},

		SMC: SMCConfig{
			BaseURL:      strings.TrimRight(getenv("SMC_BASE_URL", ""), "/"),
			WebSocketURL: getenv("SMC_WS_URL", ""),
			APIKey:       getenv("SMC_API_KEY", ""),
			APISecret:    getenv("SMC_API_SECRET", ""),
			ClientID:     getenv("SMC_CLIENT_ID", ""),
			Password:     getenv("SMC_PASSWORD", ""),
			Timeout:      getDuration("SMC_TIMEOUT", 30*time.Second),
			MaxRetries:   getInt("SMC_MAX_RETRIES", 3),
		},
	}

	return nil
}

func getenv(key, fallback string) string {

	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getCSV(key string, fallback string) []string {
	value := getenv(key, fallback)
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
