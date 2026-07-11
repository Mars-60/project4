package configs

import "time"

type AppConfig struct {
	Name    string
	Version string
	Env     string

	Server ServerConfig
	Log    LogConfig
	SMC    SMCConfig
	DB     DatabaseConfig
	Auth   AuthConfig
	Groq   GroqConfig
}

type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CORS            CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type LogConfig struct {
	Level string
}

type DatabaseConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	PasswordIterations int
}

type GroqConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	Timeout    time.Duration
	MaxRetries int
}

type SMCConfig struct {
	BaseURL      string
	WebSocketURL string
	APIKey       string
	APISecret    string
	ClientID     string
	Password     string
	Timeout      time.Duration
	MaxRetries   int
}

var App AppConfig
