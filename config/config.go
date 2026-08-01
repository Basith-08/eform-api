package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	AppName            string
	Port               string
	CORSAllowedOrigins []string
	UploadPath         string
	MaxUploadBytes     int64
	RateLimitRPS       int
	RateLimitBurst     int
	SeedOnBoot         bool
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	DBSSLMode          string
	DBTimeZone         string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		AppName:            getEnv("APP_NAME", "E-Form Employee Management System"),
		Port:               getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS"),
		UploadPath:         getEnv("UPLOAD_PATH", "./uploads"),
		MaxUploadBytes:     getEnvInt64("MAX_UPLOAD_BYTES", 1024*1024),
		RateLimitRPS:       getEnvInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst:     getEnvInt("RATE_LIMIT_BURST", 20),
		SeedOnBoot:         getEnvBool("SEED_ON_BOOT", true),
		DBHost:             getEnv("DB_HOST", "db"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBName:             getEnv("POSTGRES_DB", "eform_db"),
		DBUser:             getEnv("POSTGRES_USER", "postgres"),
		DBPassword:         getEnv("POSTGRES_PASSWORD", "postgres"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		DBTimeZone:         getEnv("DB_TIMEZONE", "UTC"),
		JWTSecret:          getEnv("JWT_SECRET", "change-this-secret"),
		AccessTokenTTL:     getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:    getEnvDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
	}
}

func (c Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode +
		" TimeZone=" + c.DBTimeZone
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}

	values := strings.Split(value, ",")
	result := make([]string, 0, len(values))
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result
}

func getEnvInt(key string, fallback int) int {
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

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
