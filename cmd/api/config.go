package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func LoadServerConfig() ServerConfig {
	return ServerConfig{
		Host:            getEnvOr("SERVER_HOST", "0.0.0.0"),
		Port:            getEnvIntOr("SERVER_PORT", 8080),
		ReadTimeout:     getEnvDurationOr("SERVER_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getEnvDurationOr("SERVER_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getEnvDurationOr("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
	}
}

func LoadDatabaseConnectionStr() string {
	host := getEnvOr("DB_HOST", "localhost")
	port := getEnvIntOr("DB_PORT", 5432)
	name := getEnvOr("DB_NAME", "swiftcodes")
	user := getEnvOr("DB_USER", "postgres")
	password := getEnvOr("DB_PASSWORD", "postgres")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		user, password, host, port, name,
	)
}

func getEnvOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvIntOr(key string, fallback int) int {
	str := os.Getenv(key)
	if str == "" {
		return fallback
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		return fallback
	}
	return v
}

func getEnvDurationOr(key string, fallback time.Duration) time.Duration {
	str := os.Getenv(key)
	if str == "" {
		return fallback
	}
	d, err := time.ParseDuration(str)
	if err != nil {
		return fallback
	}
	return d
}
