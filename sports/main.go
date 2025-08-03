package main

import (
	"database/sql"
	"flag"
	"net"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/internal/logger"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"git.neds.sh/matty/entain/sports/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	grpcEndpoint = flag.String("grpc-endpoint", "localhost:9001", "gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Initialize logger
	loggerConfig := logger.NewFromEnv()
	log, err := logger.New(loggerConfig)
	if err != nil {
		return err
	}
	defer log.Sync()

	log.Info("Starting sports service",
		zap.String("grpc_endpoint", *grpcEndpoint))

	// Initialize database connection
	database, err := sql.Open("sqlite3", "./db/sports.db")
	if err != nil {
		log.Error("Failed to open database", zap.Error(err))
		return err
	}
	defer database.Close()

	// Initialize repository
	eventsRepo := db.NewEventsRepo(database)
	if err := eventsRepo.Init(); err != nil {
		log.Error("Failed to initialize events repository", zap.Error(err))
		return err
	}

	// Initialize service
	sportsService := &service.SportsServer{
		Service: service.NewSportsService(eventsRepo, log),
	}

	// Setup gRPC server
	lis, err := net.Listen("tcp", *grpcEndpoint)
	if err != nil {
		log.Error("Failed to listen", zap.Error(err))
		return err
	}

	grpcServer := grpc.NewServer()
	sports.RegisterSportsServer(grpcServer, sportsService)

	log.Info("gRPC server listening", zap.String("address", *grpcEndpoint))

	return grpcServer.Serve(lis)
}
