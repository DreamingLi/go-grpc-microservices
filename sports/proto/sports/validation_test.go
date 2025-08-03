package sports

import (
	"strings"
	"testing"
)

func TestGetEventRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *GetEventRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid request",
			request: &GetEventRequest{Id: 1},
			wantErr: false,
		},
		{
			name:    "valid large ID",
			request: &GetEventRequest{Id: 999999},
			wantErr: false,
		},
		{
			name:    "zero ID",
			request: &GetEventRequest{Id: 0},
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name:    "negative ID",
			request: &GetEventRequest{Id: -1},
			wantErr: true,
			errMsg:  "must be positive",
		},
		{
			name:    "large negative ID",
			request: &GetEventRequest{Id: -999},
			wantErr: true,
			errMsg:  "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("GetEventRequest.Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestListEventsRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *ListEventsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil filter",
			request: &ListEventsRequest{Filter: nil},
			wantErr: false,
		},
		{
			name: "valid filter",
			request: &ListEventsRequest{
				Filter: &ListEventsRequestFilter{
					SportTypes: []string{"football", "basketball"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid filter",
			request: &ListEventsRequest{
				Filter: &ListEventsRequestFilter{
					SportTypes: []string{""},
				},
			},
			wantErr: true,
			errMsg:  "sport_types validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListEventsRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ListEventsRequest.Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestListEventsRequestFilter_ValidateSportTypes(t *testing.T) {
	tests := []struct {
		name       string
		sportTypes []string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "empty list",
			sportTypes: []string{},
			wantErr:    false,
		},
		{
			name:       "valid sport types",
			sportTypes: []string{"football", "basketball", "tennis"},
			wantErr:    false,
		},
		{
			name:       "empty sport type",
			sportTypes: []string{"football", ""},
			wantErr:    true,
			errMsg:     "empty sport type at position 1",
		},
		{
			name:       "whitespace only sport type",
			sportTypes: []string{"football", "   "},
			wantErr:    true,
			errMsg:     "empty sport type at position 1",
		},
		{
			name:       "duplicate sport types",
			sportTypes: []string{"football", "basketball", "football"},
			wantErr:    true,
			errMsg:     "duplicate sport type: football",
		},
		{
			name:       "too many sport types",
			sportTypes: make([]string, MaxSportTypes+1),
			wantErr:    true,
			errMsg:     "too many sport types",
		},
		{
			name:       "sport type too long",
			sportTypes: []string{strings.Repeat("a", MaxSportTypeLength+1)},
			wantErr:    true,
			errMsg:     "sport type too long",
		},
	}

	// Initialize the slice for the "too many sport types" test
	for i := range tests[5].sportTypes {
		tests[5].sportTypes[i] = "sport" + string(rune('0'+i%10))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &ListEventsRequestFilter{
				SportTypes: tt.sportTypes,
			}
			err := filter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListEventsRequestFilter.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ListEventsRequestFilter.Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestListEventsRequestFilter_ValidateSorting(t *testing.T) {
	tests := []struct {
		name          string
		sortField     *SortField
		sortDirection *SortDirection
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "no sorting specified",
			sortField:     nil,
			sortDirection: nil,
			wantErr:       false,
		},
		{
			name:          "valid sort field only",
			sortField:     SortField_ADVERTISED_START_TIME.Enum(),
			sortDirection: nil,
			wantErr:       false,
		},
		{
			name:          "valid sort direction only",
			sortField:     nil,
			sortDirection: SortDirection_ASC.Enum(),
			wantErr:       false,
		},
		{
			name:          "valid sort field and direction",
			sortField:     SortField_NAME.Enum(),
			sortDirection: SortDirection_DESC.Enum(),
			wantErr:       false,
		},
		{
			name:          "all valid sort fields - ADVERTISED_START_TIME",
			sortField:     SortField_ADVERTISED_START_TIME.Enum(),
			sortDirection: SortDirection_ASC.Enum(),
			wantErr:       false,
		},
		{
			name:          "all valid sort fields - NAME",
			sortField:     SortField_NAME.Enum(),
			sortDirection: SortDirection_ASC.Enum(),
			wantErr:       false,
		},
		{
			name:          "all valid sort fields - SPORT_TYPE",
			sortField:     SortField_SPORT_TYPE.Enum(),
			sortDirection: SortDirection_ASC.Enum(),
			wantErr:       false,
		},
		{
			name:          "all valid sort directions - ASC",
			sortField:     SortField_NAME.Enum(),
			sortDirection: SortDirection_ASC.Enum(),
			wantErr:       false,
		},
		{
			name:          "all valid sort directions - DESC",
			sortField:     SortField_NAME.Enum(),
			sortDirection: SortDirection_DESC.Enum(),
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &ListEventsRequestFilter{
				SortField:     tt.sortField,
				SortDirection: tt.sortDirection,
			}
			err := filter.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListEventsRequestFilter.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ListEventsRequestFilter.Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestListEventsRequestFilter_ValidateComplete(t *testing.T) {
	tests := []struct {
		name   string
		filter *ListEventsRequestFilter
		want   bool
	}{
		{
			name: "completely valid filter",
			filter: &ListEventsRequestFilter{
				SportTypes:    []string{"football", "basketball"},
				VisibleOnly:   boolPtr(true),
				SortField:     SortField_NAME.Enum(),
				SortDirection: SortDirection_ASC.Enum(),
			},
			want: false,
		},
		{
			name: "minimal valid filter",
			filter: &ListEventsRequestFilter{
				SportTypes: []string{"football"},
			},
			want: false,
		},
		{
			name: "empty filter",
			filter: &ListEventsRequestFilter{},
			want: false,
		},
		{
			name: "filter with invalid sport types",
			filter: &ListEventsRequestFilter{
				SportTypes:    []string{"football", ""},
				SortField:     SortField_NAME.Enum(),
				SortDirection: SortDirection_ASC.Enum(),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if (err != nil) != tt.want {
				t.Errorf("ListEventsRequestFilter.Validate() error = %v, want error = %v", err, tt.want)
			}
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// Benchmark tests
func BenchmarkGetEventRequest_Validate(b *testing.B) {
	req := &GetEventRequest{Id: 123}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

func BenchmarkListEventsRequest_Validate(b *testing.B) {
	req := &ListEventsRequest{
		Filter: &ListEventsRequestFilter{
			SportTypes:    []string{"football", "basketball", "tennis"},
			VisibleOnly:   boolPtr(true),
			SortField:     SortField_NAME.Enum(),
			SortDirection: SortDirection_ASC.Enum(),
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}