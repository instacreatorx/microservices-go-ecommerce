package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Env         string
	Port        string
	LogLevel    string

	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Broker   BrokerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxOpen  int
	MaxIdle  int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type BrokerConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func Load(service string) *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", service),
		Env:         getEnv("ENV", "development"),
		Port:        getEnv("PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", service),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxOpen:  getEnvInt("DB_MAX_OPEN", 25),
			MaxIdle:  getEnvInt("DB_MAX_IDLE", 5),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "super-secret-key"),
			Expiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		Broker: BrokerConfig{
			Host:     getEnv("BROKER_HOST", "localhost"),
			Port:     getEnv("BROKER_PORT", "5672"),
			User:     getEnv("BROKER_USER", "guest"),
			Password: getEnv("BROKER_PASSWORD", "guest"),
			VHost:    getEnv("BROKER_VHOST", "/"),
		},
	}
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
