package service

import (
	"context"
	"fmt"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"go.uber.org/zap"
)

// Sports defines the interface for sports-related operations.
// It provides methods to interact with sports event data and retrieve event information.
type Sports interface {
	// ListEvents retrieves a list of events based on the provided filter criteria.
	// It accepts a context for request lifecycle management and cancellation,
	// and a request containing optional filters for event visibility and sport types.
	// Returns a response with the filtered events or an error if the operation fails.
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)

	// GetEvent retrieves a single event by its ID.
	// It accepts a context for request lifecycle management and cancellation,
	// and a request containing the event ID to retrieve.
	// Returns a response with the event or an error if the operation fails.
	GetEvent(ctx context.Context, in *sports.GetEventRequest) (*sports.GetEventResponse, error)
}

type sportsService struct {
	eventsRepo db.EventsRepo
	logger     *zap.Logger
}

// NewSportsService creates a new sports service with injected logger
func NewSportsService(eventsRepo db.EventsRepo, logger *zap.Logger) Sports {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &sportsService{
		eventsRepo: eventsRepo,
		logger:     logger,
	}
}

func (s *sportsService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	reqLogger := s.logger.With(
		zap.String("method", "ListEvents"),
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

	reqLogger.Debug("Calling repository")

	// Call repository
	events, err := s.eventsRepo.List(in.Filter)
	if err != nil {
		reqLogger.Error("Repository call failed",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	return &sports.ListEventsResponse{Events: events}, nil
}

func (s *sportsService) GetEvent(ctx context.Context, in *sports.GetEventRequest) (*sports.GetEventResponse, error) {
	reqLogger := s.logger.With(
		zap.String("method", "GetEvent"),
		zap.Int64("event_id", in.GetId()),
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
		reqLogger.Warn("Request validation failed: invalid event ID",
			zap.Int64("event_id", in.Id),
		)
		return nil, fmt.Errorf("event ID must be greater than 0")
	}

	reqLogger.Debug("Calling repository")

	// Call repository
	event, err := s.eventsRepo.GetByID(in.Id)
	if err != nil {
		reqLogger.Error("Repository call failed",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve event: %w", err)
	}

	return &sports.GetEventResponse{Event: event}, nil
}

// SportsServer is a gRPC server wrapper that embeds the required UnimplementedSportsServer
type SportsServer struct {
	sports.UnimplementedSportsServer
	Service Sports
}

// ListEvents implements the gRPC SportsServer interface
func (s *SportsServer) ListEvents(ctx context.Context, req *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	return s.Service.ListEvents(ctx, req)
}

// GetEvent implements the gRPC SportsServer interface
func (s *SportsServer) GetEvent(ctx context.Context, req *sports.GetEventRequest) (*sports.GetEventResponse, error) {
	return s.Service.GetEvent(ctx, req)
}