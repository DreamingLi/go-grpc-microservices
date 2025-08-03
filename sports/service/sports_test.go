package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/testing/protocmp"
)

// testEventsRepo is a simple mock implementation for testing
type testEventsRepo struct {
	events     []*sports.Event
	err        error
	lastFilter *sports.ListEventsRequestFilter
	initCalled bool
}

// GetByID implements the db.EventsRepo interface for testing.
func (t *testEventsRepo) GetByID(id int64) (*sports.Event, error) {
	for _, event := range t.events {
		if event.Id == id {
			return event, nil
		}
	}
	return nil, errors.New("event not found")
}

// List implements the db.EventsRepo interface for testing.
func (t *testEventsRepo) List(filter *sports.ListEventsRequestFilter) ([]*sports.Event, error) {
	t.lastFilter = filter
	if t.err != nil {
		return nil, t.err
	}
	return t.events, nil
}

func (t *testEventsRepo) Init() error {
	t.initCalled = true
	return t.err
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create test repository
func newTestEventsRepo(events []*sports.Event, err error) db.EventsRepo {
	return &testEventsRepo{
		events: events,
		err:    err,
	}
}

func TestNewSportsService(t *testing.T) {
	tests := []struct {
		name   string
		repo   db.EventsRepo
		logger *zap.Logger
	}{
		{
			name:   "with logger",
			repo:   newTestEventsRepo(nil, nil),
			logger: zaptest.NewLogger(t),
		},
		{
			name:   "with nil logger",
			repo:   newTestEventsRepo(nil, nil),
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSportsService(tt.repo, tt.logger)
			if service == nil {
				t.Error("NewSportsService() = nil, want non-nil service")
			}
		})
	}
}

func TestSportsService_ListEvents_Success(t *testing.T) {
	testEvents := []*sports.Event{
		{
			Id:        1,
			Name:      "Team A vs Team B",
			SportType: "football",
			Venue:     "Stadium A",
			Visible:   true,
		},
		{
			Id:        2,
			Name:      "Team C vs Team D",
			SportType: "basketball",
			Venue:     "Arena B",
			Visible:   false,
		},
	}

	repo := newTestEventsRepo(testEvents, nil)
	logger := zaptest.NewLogger(t)
	service := NewSportsService(repo, logger)

	request := &sports.ListEventsRequest{
		Filter: &sports.ListEventsRequestFilter{
			VisibleOnly: boolPtr(true),
		},
	}

	response, err := service.ListEvents(context.Background(), request)

	if err != nil {
		t.Errorf("ListEvents() error = %v, want nil", err)
	}

	if response == nil {
		t.Error("ListEvents() response = nil, want non-nil")
		return
	}

	if diff := cmp.Diff(testEvents, response.Events, protocmp.Transform()); diff != "" {
		t.Errorf("ListEvents() events mismatch (-want +got):\n%s", diff)
	}
}

func TestSportsService_GetEvent_Success(t *testing.T) {
	testEvent := &sports.Event{
		Id:        1,
		Name:      "Championship Match",
		SportType: "football",
		Venue:     "Stadium A",
		Visible:   true,
	}

	repo := newTestEventsRepo([]*sports.Event{testEvent}, nil)
	logger := zaptest.NewLogger(t)
	service := NewSportsService(repo, logger)

	request := &sports.GetEventRequest{Id: 1}

	response, err := service.GetEvent(context.Background(), request)

	if err != nil {
		t.Errorf("GetEvent() error = %v, want nil", err)
	}

	if response == nil {
		t.Error("GetEvent() response = nil, want non-nil")
		return
	}

	if diff := cmp.Diff(testEvent, response.Event, protocmp.Transform()); diff != "" {
		t.Errorf("GetEvent() event mismatch (-want +got):\n%s", diff)
	}
}

func TestSportsService_GetEvent_NotFound(t *testing.T) {
	repo := newTestEventsRepo([]*sports.Event{}, nil)
	logger := zaptest.NewLogger(t)
	service := NewSportsService(repo, logger)

	request := &sports.GetEventRequest{Id: 999}

	response, err := service.GetEvent(context.Background(), request)

	if err == nil {
		t.Error("GetEvent() with non-existent ID error = nil, want error")
		return
	}

	wantErrorMsg := "failed to retrieve event"
	if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("GetEvent() error = %v, want error containing %q", err, wantErrorMsg)
	}

	if response != nil {
		t.Errorf("GetEvent() with non-existent ID response = %v, want nil", response)
	}
}

func TestSportsService_ListEvents_NilRequest(t *testing.T) {
	repo := newTestEventsRepo(nil, nil)
	logger := zaptest.NewLogger(t)
	service := NewSportsService(repo, logger)

	response, err := service.ListEvents(context.Background(), nil)

	if err == nil {
		t.Error("ListEvents() with nil request error = nil, want error")
		return
	}

	wantErrorMsg := "request cannot be nil"
	if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("ListEvents() error = %v, want error containing %q", err, wantErrorMsg)
	}

	if response != nil {
		t.Errorf("ListEvents() with nil request response = %v, want nil", response)
	}
}

func TestSportsService_GetEvent_InvalidID(t *testing.T) {
	repo := newTestEventsRepo(nil, nil)
	logger := zaptest.NewLogger(t)
	service := NewSportsService(repo, logger)

	tests := []struct {
		name string
		id   int64
	}{
		{"zero ID", 0},
		{"negative ID", -1},
		{"negative large ID", -999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &sports.GetEventRequest{Id: tt.id}

			response, err := service.GetEvent(context.Background(), request)

			if err == nil {
				t.Errorf("GetEvent() with invalid ID %d error = nil, want error", tt.id)
				return
			}

			wantErrorMsg := "invalid request"
			if !strings.Contains(err.Error(), wantErrorMsg) {
				t.Errorf("GetEvent() error = %v, want error containing %q", err, wantErrorMsg)
			}

			if response != nil {
				t.Errorf("GetEvent() with invalid ID response = %v, want nil", response)
			}
		})
	}
}

// Benchmark test for ListEvents
func BenchmarkSportsService_ListEvents(b *testing.B) {
	events := make([]*sports.Event, 100)
	for i := 0; i < 100; i++ {
		events[i] = &sports.Event{
			Id:        int64(i + 1),
			Name:      "Benchmark Event",
			SportType: "football",
			Venue:     "Stadium A",
			Visible:   i%2 == 0,
		}
	}

	repo := newTestEventsRepo(events, nil)
	logger := zap.NewNop() // Use no-op logger for benchmarks
	service := NewSportsService(repo, logger)
	request := &sports.ListEventsRequest{
		Filter: &sports.ListEventsRequestFilter{
			VisibleOnly: boolPtr(true),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ListEvents(context.Background(), request)
		if err != nil {
			b.Fatalf("ListEvents() failed: %v", err)
		}
	}
}
