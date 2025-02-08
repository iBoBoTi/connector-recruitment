package config

import (
	"log"
	"os"
	"strconv"
)

type DBConfig struct {
	Host           string
	Port           int
	User           string
	Password       string
	Name           string
	MigrationsPath string
}

type GRPCServerConfig struct {
	Port string
}

type AWSConfig struct {
	Endpoint string
	Region   string
}

type Config struct {
	DB         DBConfig
	GRPCServer GRPCServerConfig
	AWS        AWSConfig
}

// LoadConfig loads configuration from environment variables or defaults.
func LoadConfig() *Config {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	return &Config{
		DB: DBConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           port,
			User:           getEnv("DB_USER", "aryon"),
			Password:       getEnv("DB_PASS", "aryon"),
			Name:           getEnv("DB_NAME", "aryondb"),
			MigrationsPath: getEnv("DB_MIGRATIONS_PATH", "migrations"),
		},
		GRPCServer: GRPCServerConfig{
			Port: getEnv("GRPC_SERVER_PORT", "50051"),
		},
		AWS: AWSConfig{
			Endpoint: getEnv("AWS_ENDPOINT", "http://localhost:4566"),
			Region:   getEnv("AWS_REGION", "us-east-1"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
