package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/testing/protocmp"
)

// testRacesRepo is a simple mock implementation for testing
type testRacesRepo struct {
	races      []*racing.Race
	err        error
	lastFilter *racing.ListRacesRequestFilter
	initCalled bool
	delay      time.Duration // Add delay for testing slow queries
}

func (t *testRacesRepo) Init() error {
	t.initCalled = true
	return t.err
}

func (t *testRacesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	t.lastFilter = filter

	// Simulate delay if configured
	if t.delay > 0 {
		time.Sleep(t.delay)
	}

	if t.err != nil {
		return nil, t.err
	}
	return t.races, nil
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Helper function to create test repository
func newTestRepo(races []*racing.Race, err error) db.RacesRepo {
	return &testRacesRepo{
		races: races,
		err:   err,
	}
}

func TestNewRacingService(t *testing.T) {
	tests := []struct {
		name   string
		repo   db.RacesRepo
		logger *zap.Logger
	}{
		{
			name:   "with logger",
			repo:   newTestRepo(nil, nil),
			logger: zaptest.NewLogger(t),
		},
		{
			name:   "with nil logger",
			repo:   newTestRepo(nil, nil),
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewRacingService(tt.repo, tt.logger)
			if service == nil {
				t.Error("NewRacingService() = nil, want non-nil service")
			}
		})
	}
}

func TestRacingService_ListRaces_Success(t *testing.T) {
	testRaces := []*racing.Race{
		{
			Id:        1,
			MeetingId: 1,
			Name:      "Test Race 1",
			Number:    1,
			Visible:   true,
		},
		{
			Id:        2,
			MeetingId: 2,
			Name:      "Test Race 2",
			Number:    2,
			Visible:   false,
		},
	}

	repo := newTestRepo(testRaces, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			VisibleOnly: boolPtr(true),
		},
	}

	response, err := service.ListRaces(context.Background(), request)

	if err != nil {
		t.Errorf("ListRaces() error = %v, want nil", err)
	}

	if response == nil {
		t.Error("ListRaces() response = nil, want non-nil")
		return
	}

	if diff := cmp.Diff(testRaces, response.Races, protocmp.Transform()); diff != "" {
		t.Errorf("ListRaces() races mismatch (-want +got):\n%s", diff)
	}
}

func TestRacingService_ListRaces_NilRequest(t *testing.T) {
	repo := newTestRepo(nil, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	response, err := service.ListRaces(context.Background(), nil)

	if err == nil {
		t.Error("ListRaces() with nil request error = nil, want error")
		return
	}

	wantErrorMsg := "request cannot be nil"
	if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("ListRaces() error = %v, want error containing %q", err, wantErrorMsg)
	}

	if response != nil {
		t.Errorf("ListRaces() with nil request response = %v, want nil", response)
	}
}

func TestRacingService_ListRaces_CancelledContext(t *testing.T) {
	repo := newTestRepo(nil, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{},
	}

	response, err := service.ListRaces(ctx, request)

	if err == nil {
		t.Error("ListRaces() with cancelled context error = nil, want error")
	}

	wantErrorMsg := "request cancelled"
	if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("ListRaces() error = %v, want error containing %q", err, wantErrorMsg)
	}

	if response != nil {
		t.Errorf("ListRaces() with cancelled context response = %v, want nil", response)
	}
}

func TestRacingService_ListRaces_NilFilter(t *testing.T) {
	testRaces := []*racing.Race{
		{Id: 1, Name: "Race 1", Visible: true},
	}

	repo := newTestRepo(testRaces, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: nil,
	}

	response, err := service.ListRaces(context.Background(), request)

	if err != nil {
		t.Errorf("ListRaces() with nil filter error = %v, want nil", err)
	}

	if response == nil {
		t.Error("ListRaces() with nil filter response = nil, want non-nil")
		return
	}

	if len(response.Races) != len(testRaces) {
		t.Errorf("ListRaces() with nil filter returned %d races, want %d",
			len(response.Races), len(testRaces))
	}
}

func TestRacingService_ListRaces_EmptyFilter(t *testing.T) {
	testRaces := []*racing.Race{
		{Id: 1, Name: "Race 1", Visible: true},
		{Id: 2, Name: "Race 2", Visible: false},
	}

	repo := newTestRepo(testRaces, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{},
	}

	response, err := service.ListRaces(context.Background(), request)

	if err != nil {
		t.Errorf("ListRaces() with empty filter error = %v, want nil", err)
	}

	if response == nil {
		t.Error("ListRaces() with empty filter response = nil, want non-nil")
		return
	}

	if len(response.Races) != len(testRaces) {
		t.Errorf("ListRaces() with empty filter returned %d races, want %d",
			len(response.Races), len(testRaces))
	}
}

func TestRacingService_ListRaces_EmptyResults(t *testing.T) {
	emptyRaces := []*racing.Race{}

	repo := newTestRepo(emptyRaces, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			VisibleOnly: boolPtr(true),
			MeetingIds:  []int64{999},
		},
	}

	response, err := service.ListRaces(context.Background(), request)

	if err != nil {
		t.Errorf("ListRaces() with non-existent meeting error = %v, want nil", err)
	}

	if response == nil {
		t.Error("ListRaces() with non-existent meeting response = nil, want non-nil")
		return
	}

	if len(response.Races) != 0 {
		t.Errorf("ListRaces() with non-existent meeting returned %d races, want 0",
			len(response.Races))
	}
}

func TestRacingService_ListRaces_RepositoryError(t *testing.T) {
	expectedError := errors.New("database connection failed")
	repo := newTestRepo(nil, expectedError)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			VisibleOnly: boolPtr(true),
		},
	}

	response, err := service.ListRaces(context.Background(), request)

	if err == nil {
		t.Error("ListRaces() with repository error = nil, want error")
		return
	}

	wantErrorMsg := "failed to retrieve races"
	if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("ListRaces() error = %v, want error containing %q", err, wantErrorMsg)
	}

	if !strings.Contains(err.Error(), expectedError.Error()) {
		t.Errorf("ListRaces() error = %v, want error containing %q", err, expectedError.Error())
	}

	if response != nil {
		t.Errorf("ListRaces() with repository error response = %v, want nil", response)
	}
}

func TestRacingService_ListRaces_FilterPropagation(t *testing.T) {
	tests := []struct {
		name   string
		filter *racing.ListRacesRequestFilter
	}{
		{
			name:   "nil filter",
			filter: nil,
		},
		{
			name:   "empty filter",
			filter: &racing.ListRacesRequestFilter{},
		},
		{
			name: "visible only true",
			filter: &racing.ListRacesRequestFilter{
				VisibleOnly: boolPtr(true),
			},
		},
		{
			name: "visible only false",
			filter: &racing.ListRacesRequestFilter{
				VisibleOnly: boolPtr(false),
			},
		},
		{
			name: "meeting ids only",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds: []int64{1, 2, 3},
			},
		},
		{
			name: "combined filters",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds:  []int64{1, 2},
				VisibleOnly: boolPtr(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRepo := &testRacesRepo{
				races: []*racing.Race{},
				err:   nil,
			}
			logger := zaptest.NewLogger(t)
			service := NewRacingService(testRepo, logger)

			request := &racing.ListRacesRequest{Filter: tt.filter}

			_, err := service.ListRaces(context.Background(), request)
			if err != nil {
				t.Errorf("ListRaces() failed: %v", err)
				return
			}

			if diff := cmp.Diff(tt.filter, testRepo.lastFilter, protocmp.Transform()); diff != "" {
				t.Errorf("Filter propagation mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRacingService_ListRaces_ResponseValidation(t *testing.T) {
	testRaces := []*racing.Race{
		{
			Id:        1,
			MeetingId: 100,
			Name:      "Validation Test Race",
			Number:    5,
			Visible:   true,
		},
	}

	repo := newTestRepo(testRaces, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{},
	}

	response, err := service.ListRaces(context.Background(), request)

	if err != nil {
		t.Fatalf("ListRaces() failed: %v", err)
	}

	if response == nil {
		t.Fatal("ListRaces() returned nil response")
	}

	if len(response.Races) != 1 {
		t.Fatalf("ListRaces() returned %d races, want 1", len(response.Races))
	}

	race := response.Races[0]

	if race.Id != 1 {
		t.Errorf("Race ID = %d, want 1", race.Id)
	}
	if race.MeetingId != 100 {
		t.Errorf("Race MeetingId = %d, want 100", race.MeetingId)
	}
	if race.Name != "Validation Test Race" {
		t.Errorf("Race Name = %q, want %q", race.Name, "Validation Test Race")
	}
	if race.Number != 5 {
		t.Errorf("Race Number = %d, want 5", race.Number)
	}
	if !race.Visible {
		t.Errorf("Race Visible = %t, want true", race.Visible)
	}
}

func TestRacingService_ListRaces_ValidationError(t *testing.T) {
	repo := newTestRepo(nil, nil)
	logger := zaptest.NewLogger(t)
	service := NewRacingService(repo, logger)

	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			MeetingIds: []int64{10, 20, 10}, // '10' is a duplicate ID
		},
	}

	// The validation should now fail
	if request.Validate() == nil {
		t.Fatal("Expected validation to fail, but it passed.")
	}

	response, err := service.ListRaces(context.Background(), request)

	if err == nil {
		t.Error("ListRaces() with invalid request error = nil, want error")
	}

	wantErrorMsg := "validation failed"
	if !strings.Contains(err.Error(), wantErrorMsg) {
		t.Errorf("ListRaces() error = %v, want error containing %q", err, wantErrorMsg)
	}

	if response != nil {
		t.Errorf("ListRaces() with invalid request response = %v, want nil", response)
	}
}

// Benchmark test
func BenchmarkRacingService_ListRaces(b *testing.B) {
	races := make([]*racing.Race, 100)
	for i := 0; i < 100; i++ {
		races[i] = &racing.Race{
			Id:        int64(i + 1),
			MeetingId: int64((i % 10) + 1),
			Name:      "Benchmark Race",
			Number:    int64(i + 1),
			Visible:   i%2 == 0,
		}
	}

	repo := newTestRepo(races, nil)
	logger := zap.NewNop() // Use no-op logger for benchmarks
	service := NewRacingService(repo, logger)
	request := &racing.ListRacesRequest{
		Filter: &racing.ListRacesRequestFilter{
			VisibleOnly: boolPtr(true),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ListRaces(context.Background(), request)
		if err != nil {
			b.Fatalf("ListRaces() failed: %v", err)
		}
	}
}

// Table-driven test for multiple scenarios
func TestRacingService_ListRaces_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		races         []*racing.Race
		repoError     error
		request       *racing.ListRacesRequest
		ctx           context.Context
		wantError     bool
		wantRaces     int
		errorContains string
	}{
		{
			name:      "successful request",
			races:     []*racing.Race{{Id: 1, Name: "Race 1", Visible: true}},
			repoError: nil,
			request:   &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}},
			ctx:       context.Background(),
			wantError: false,
			wantRaces: 1,
		},
		{
			name:          "nil request",
			races:         nil,
			repoError:     nil,
			request:       nil,
			ctx:           context.Background(),
			wantError:     true,
			wantRaces:     0,
			errorContains: "request cannot be nil",
		},
		{
			name:          "nil context",
			races:         nil,
			repoError:     nil,
			request:       &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}},
			ctx:           nil,
			wantError:     true,
			wantRaces:     0,
			errorContains: "context cannot be nil",
		},
		{
			name:          "repository error",
			races:         nil,
			repoError:     errors.New("db error"),
			request:       &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}},
			ctx:           context.Background(),
			wantError:     true,
			wantRaces:     0,
			errorContains: "failed to retrieve races",
		},
		{
			name:      "empty results",
			races:     []*racing.Race{},
			repoError: nil,
			request:   &racing.ListRacesRequest{Filter: &racing.ListRacesRequestFilter{}},
			ctx:       context.Background(),
			wantError: false,
			wantRaces: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(tt.races, tt.repoError)
			logger := zaptest.NewLogger(t)
			service := NewRacingService(repo, logger)

			response, err := service.ListRaces(tt.ctx, tt.request)

			if tt.wantError {
				if err == nil {
					t.Errorf("ListRaces() error = nil, want error")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ListRaces() error = %v, want error containing %q", err, tt.errorContains)
				}
				if response != nil {
					t.Errorf("ListRaces() response = %v, want nil", response)
				}
			} else {
				if err != nil {
					t.Errorf("ListRaces() error = %v, want nil", err)
				}
				if response == nil {
					t.Error("ListRaces() response = nil, want non-nil")
					return
				}
				if len(response.Races) != tt.wantRaces {
					t.Errorf("ListRaces() returned %d races, want %d", len(response.Races), tt.wantRaces)
				}
			}
		})
	}
}
