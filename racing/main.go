package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/internal/logger"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"git.neds.sh/matty/entain/racing/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	grpcEndpoint = flag.String("grpc-endpoint", "localhost:9000", "gRPC server endpoint")
)

func main() {
	flag.Parse()

	// 1. init logger
	loggerConfig := logger.NewFromEnv()
	serviceLogger, err := logger.New(loggerConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := serviceLogger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
		}
	}()

	serviceLogger.Info("Starting racing service",
		zap.String("endpoint", *grpcEndpoint),
		zap.String("log_level", string(loggerConfig.Level)),
		zap.String("environment", string(loggerConfig.Environment)),
	)

	// 2. run the service
	if err := run(serviceLogger); err != nil {
		serviceLogger.Fatal("Service failed to start", zap.Error(err))
	}
}

func run(logger *zap.Logger) error {
	logger.Info("Initializing gRPC server")

	conn, err := net.Listen("tcp", ":9000")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.Info("Setting up database connection")
	racingDB, err := sql.Open("sqlite3", "./db/racing.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer racingDB.Close()

	logger.Info("Initializing repository")
	racesRepo := db.NewRacesRepo(racingDB)
	if err := racesRepo.Init(); err != nil {
		logger.Error("Failed to initialize repository", zap.Error(err))
		return fmt.Errorf("failed to initialize repository: %w", err)
	}

	// 3. create acing service，inject logger
	logger.Info("Creating racing service")
	racingService := service.NewRacingService(racesRepo, logger)

	logger.Info("Setting up gRPC server")
	grpcServer := grpc.NewServer()

	racing.RegisterRacingServer(grpcServer, racingService)

	logger.Info("gRPC server listening", zap.String("address", *grpcEndpoint))

	if err := grpcServer.Serve(conn); err != nil {
		logger.Error("gRPC server failed", zap.Error(err))
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}
