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
	port, err := strconv.Atoi(GetEnv("DB_PORT", "5432"))
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	return &Config{
		DB: DBConfig{
			Host:           GetEnv("DB_HOST", "localhost"),
			Port:           port,
			User:           GetEnv("DB_USER", "aryon"),
			Password:       GetEnv("DB_PASS", "aryon"),
			Name:           GetEnv("DB_NAME", "aryondb"),
			MigrationsPath: GetEnv("DB_MIGRATIONS_PATH", "migrations"),
		},
		GRPCServer: GRPCServerConfig{
			Port: GetEnv("GRPC_SERVER_PORT", "50051"),
		},
		AWS: AWSConfig{
			Endpoint: GetEnv("AWS_ENDPOINT", "http://localhost:4566"),
			Region:   GetEnv("AWS_REGION", "us-east-1"),
		},
	}
}

func GetEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
