package service

import (
	"context"
	"fmt"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"go.uber.org/zap"
)

// Racing defines the interface for racing-related operations.
// It provides methods to interact with racing data and retrieve race information.
type Racing interface {
	// ListRaces retrieves a list of races based on the provided filter criteria.
	// It accepts a context for request lifecycle management and cancellation,
	// and a request containing optional filters for race visibility and meeting IDs.
	// Returns a response with the filtered races or an error if the operation fails.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)

	// GetRace retrieves a single race by its ID.
	// It accepts a context for request lifecycle management and cancellation,
	// and a request containing the race ID to retrieve.
	// Returns a response with the race or an error if the operation fails.
	GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.GetRaceResponse, error)
}

type racingService struct {
	racesRepo db.RacesRepo
	logger    *zap.Logger
}

// NewRacingService creates a new racing service with injected logger
func NewRacingService(racesRepo db.RacesRepo, logger *zap.Logger) Racing {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &racingService{
		racesRepo: racesRepo,
		logger:    logger,
	}
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	reqLogger := s.logger.With(
		zap.String("method", "ListRaces"),
	)

	reqLogger.Debug("Request started", zap.Any("filter", in.GetFilter()))

	// Context validation
	if ctx == nil {
		reqLogger.Error("Context validation failed: nil context")
		return nil, fmt.Errorf("context cannot be nil")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		reqLogger.Warn("Request cancelled",
			zap.Error(ctx.Err()),
		)
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	default:
		// Continue processing
	}

	// Input validation
	if in == nil {
		reqLogger.Warn("Request validation failed: nil request")
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request
	if err := in.Validate(); err != nil {
		reqLogger.Warn("Request validation failed",
			zap.Error(err),
			zap.Any("filter", in.Filter),
		)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	reqLogger.Debug("Calling repository")

	// Call repository
	races, err := s.racesRepo.List(in.Filter)
	if err != nil {
		reqLogger.Error("Repository call failed",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve races: %w", err)
	}

	return &racing.ListRacesResponse{Races: races}, nil
}

func (s *racingService) GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.GetRaceResponse, error) {
	reqLogger := s.logger.With(
		zap.String("method", "GetRace"),
		zap.Int64("race_id", in.GetId()),
	)

	reqLogger.Debug("Request started")

	// Context validation
	if ctx == nil {
		reqLogger.Error("Context validation failed: nil context")
		return nil, fmt.Errorf("context cannot be nil")
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		reqLogger.Warn("Request cancelled",
			zap.Error(ctx.Err()),
		)
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	default:
		// Continue processing
	}

	// Input validation
	if in == nil {
		reqLogger.Warn("Request validation failed: nil request")
		return nil, fmt.Errorf("request cannot be nil")
	}

	if in.Id <= 0 {
		reqLogger.Warn("Request validation failed: invalid race ID",
			zap.Int64("race_id", in.Id),
		)
		return nil, fmt.Errorf("race ID must be greater than 0")
	}

	reqLogger.Debug("Calling repository")

	// Call repository
	race, err := s.racesRepo.GetByID(in.Id)
	if err != nil {
		reqLogger.Error("Repository call failed",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve race: %w", err)
	}

	return &racing.GetRaceResponse{Race: race}, nil
}
