package main

import (
	"context"
	"embed"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pressly/goose/v3"

	"github.com/iBoBoTi/connector-service/config"
	connector_v1 "github.com/iBoBoTi/connector-service/gen/proto"
	"github.com/iBoBoTi/connector-service/internal/repository"
	"github.com/iBoBoTi/connector-service/internal/services"
	handler "github.com/iBoBoTi/connector-service/internal/transport/grpc"
	"github.com/iBoBoTi/connector-service/internal/usecase"
	"github.com/iBoBoTi/connector-service/pkg/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load configuration
	cfg := config.LoadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting application")

	// Initialize DB connection
	dbConn, err := db.NewPostgresDB(cfg.DB)
	if err != nil {
		slog.Error("Failed to open DB", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	goose.SetBaseFS(embedMigrations)
	if err := goose.Up(dbConn, "migrations"); err != nil {
		slog.Error("Failed to apply migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("Migrations applied successfully")

	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		Region:           aws.String(cfg.AWS.Region),
		Endpoint:         aws.String(cfg.AWS.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		slog.Error("Failed to create AWS session", "error", err)
		os.Exit(1)
	}

	// Setup repository, clients, and usecase
	connRepo := repository.NewConnectorRepository(dbConn)
	secretsClient := services.NewSecretsManager(sess)
	slackClient := services.NewSlackClient()
	connUsecase := usecase.NewConnectorUsecase(connRepo, secretsClient, slackClient)
	connHandler := handler.NewSlackConnectorHandler(connUsecase)

	// Create and register gRPC server
	grpcServer := grpc.NewServer()
	connector_v1.RegisterSlackConnectorServiceServer(grpcServer, connHandler)
	reflection.Register(grpcServer)

	// Listen on the desired port
	grpcAddr := cfg.GRPCServer.Port
	listener, err := net.Listen("tcp", ":"+grpcAddr)
	if err != nil {
		slog.Error("Failed to listen on gRPC port", "port", grpcAddr, "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("gRPC server is running", "port", grpcAddr)
		if serveErr := grpcServer.Serve(listener); serveErr != nil {
			slog.Error("gRPC server encountered an error", "error", serveErr)
			stop()
		}
	}()

	<-ctx.Done()

	slog.Info("Shutting down gracefully...")
	grpcServer.GracefulStop()

	time.Sleep(1 * time.Second)
	slog.Info("Server stopped. Goodbye.")
}
