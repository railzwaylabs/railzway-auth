package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config contains runtime configuration values.
type Config struct {
	Environment          string
	HTTPPort             string
	DBHost               string
	DBPort               string
	DBName               string
	DBUser               string
	DBPassword           string
	DBSSLMode            string
	AdminEmail           string
	AdminPassword        string
	DefaultOrgID         int64
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	RefreshTokenBytes    int
	ServiceName          string
	RateLimitRPM         int
	OTLPEndpoint         string
	OTLPInsecure         bool
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	CORSAllowCredentials bool
}

// DSN returns the database connection string.
func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

// Load reads configuration from environment variables with sane defaults.
func Load() (Config, error) {
	_ = godotenv.Load()

	defaultOrgRaw := strings.TrimSpace(os.Getenv("DEFAULT_ORG"))
	if defaultOrgRaw == "" {
		return Config{}, fmt.Errorf("DEFAULT_ORG is required")
	}
	defaultOrgID, err := strconv.ParseInt(defaultOrgRaw, 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("DEFAULT_ORG must be a valid int64")
	}

	cfg := Config{
		Environment:          getEnv("APP_ENV", "development"),
		HTTPPort:             getEnv("HTTP_PORT", "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBName:               getEnv("DB_NAME", "postgres"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		DBSSLMode:            getEnv("DB_SSL_MODE", "disable"),
		DefaultOrgID:         defaultOrgID,
		RedisAddr:            getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:        os.Getenv("REDIS_PASSWORD"),
		RedisDB:              getInt("REDIS_DB", 0),
		AccessTokenTTL:       getDuration("ACCESS_TOKEN_TTL", time.Hour),
		RefreshTokenTTL:      getDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		RefreshTokenBytes:    getInt("REFRESH_TOKEN_BYTES", 32),
		ServiceName:          getEnv("SERVICE_NAME", "railzway-auth"),
		RateLimitRPM:         getInt("RATE_LIMIT_RPM", 600),
		OTLPEndpoint:         os.Getenv("OTLP_ENDPOINT"),
		OTLPInsecure:         getBool("OTLP_INSECURE", true),
		CORSAllowedOrigins:   getList("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSAllowedMethods:   getList("CORS_ALLOWED_METHODS", []string{"GET", "POST", "OPTIONS"}),
		CORSAllowedHeaders:   getList("CORS_ALLOWED_HEADERS", []string{"Authorization", "Content-Type"}),
		CORSAllowCredentials: getBool("CORS_ALLOW_CREDENTIALS", false),
	}

	if cfg.RefreshTokenBytes < 32 {
		cfg.RefreshTokenBytes = 32
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return def
}

func getInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		switch strings.ToLower(v) {
		case "1", "true", "t", "yes", "y", "on":
			return true
		case "0", "false", "f", "no", "n", "off":
			return false
		}
	}
	return def
}

func getList(key string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok {
		parts := strings.Split(v, ",")
		var cleaned []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cleaned = append(cleaned, trimmed)
			}
		}
		if len(cleaned) > 0 {
			return cleaned
		}
	}
	return def
}
