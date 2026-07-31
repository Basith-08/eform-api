package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	AppName         string
	Port            string
	FrontendURL     string
	UploadPath      string
	MaxUploadBytes  int64
	RateLimitRPS    int
	RateLimitBurst  int
	SeedOnBoot      bool
	DBHost          string
	DBPort          string
	DBName          string
	DBUser          string
	DBPassword      string
	DBSSLMode       string
	DBTimeZone      string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() Config {
	return Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppName:         getEnv("APP_NAME", "E-Form Employee Management System"),
		Port:            getEnv("PORT", "8080"),
		FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:5173"),
		UploadPath:      getEnv("UPLOAD_PATH", "./uploads"),
		MaxUploadBytes:  getEnvInt64("MAX_UPLOAD_BYTES", 1024*1024),
		RateLimitRPS:    getEnvInt("RATE_LIMIT_RPS", 10),
		RateLimitBurst:  getEnvInt("RATE_LIMIT_BURST", 20),
		SeedOnBoot:      getEnvBool("SEED_ON_BOOT", true),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBName:          getEnv("DB_NAME", "eform"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		DBTimeZone:      getEnv("DB_TIMEZONE", "UTC"),
		JWTSecret:       getEnv("JWT_SECRET", "change-this-secret"),
		AccessTokenTTL:  getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getEnvDuration("REFRESH_TOKEN_TTL", 7*24*time.Hour),
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
