package racing

import (
	"strings"
	"testing"
)

func TestListRacesRequestFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  *ListRacesRequestFilter
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil filter is valid",
			filter:  nil,
			wantErr: false,
		},
		{
			name:    "empty filter is valid",
			filter:  &ListRacesRequestFilter{},
			wantErr: false,
		},
		{
			name: "valid meeting ids",
			filter: &ListRacesRequestFilter{
				MeetingIds: []int64{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "negative meeting id",
			filter: &ListRacesRequestFilter{
				MeetingIds: []int64{1, -2, 3},
			},
			wantErr: true,
			errMsg:  "invalid meeting ID at position 1: -2",
		},
		{
			name: "zero meeting id",
			filter: &ListRacesRequestFilter{
				MeetingIds: []int64{0},
			},
			wantErr: true,
			errMsg:  "invalid meeting ID at position 0: 0",
		},
		{
			name: "duplicate meeting ids",
			filter: &ListRacesRequestFilter{
				MeetingIds: []int64{1, 2, 1},
			},
			wantErr: true,
			errMsg:  "duplicate meeting ID: 1",
		},
		{
			name: "meeting id too large",
			filter: &ListRacesRequestFilter{
				MeetingIds: []int64{1000000},
			},
			wantErr: true,
			errMsg:  "meeting ID too large",
		},
		{
			name: "valid visible only",
			filter: &ListRacesRequestFilter{
				VisibleOnly: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "combined valid filters",
			filter: &ListRacesRequestFilter{
				MeetingIds:  []int64{1, 2, 3},
				VisibleOnly: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "valid sort field - ADVERTISED_START_TIME",
			filter: &ListRacesRequestFilter{
				SortField: sortFieldPtr(SortField_ADVERTISED_START_TIME),
			},
			wantErr: false,
		},
		{
			name: "valid sort field - NAME",
			filter: &ListRacesRequestFilter{
				SortField: sortFieldPtr(SortField_NAME),
			},
			wantErr: false,
		},
		{
			name: "valid sort field - NUMBER",
			filter: &ListRacesRequestFilter{
				SortField: sortFieldPtr(SortField_NUMBER),
			},
			wantErr: false,
		},
		{
			name: "valid sort direction - ASC",
			filter: &ListRacesRequestFilter{
				SortDirection: sortDirectionPtr(SortDirection_ASC),
			},
			wantErr: false,
		},
		{
			name: "valid sort direction - DESC",
			filter: &ListRacesRequestFilter{
				SortDirection: sortDirectionPtr(SortDirection_DESC),
			},
			wantErr: false,
		},
		{
			name: "valid complete filter with sorting",
			filter: &ListRacesRequestFilter{
				MeetingIds:    []int64{1, 2, 3},
				VisibleOnly:   boolPtr(true),
				SortField:     sortFieldPtr(SortField_NAME),
				SortDirection: sortDirectionPtr(SortDirection_DESC),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.filter != nil {
				err = tt.filter.Validate()
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// sortFieldPtr returns a pointer to the given SortField value
func sortFieldPtr(sf SortField) *SortField {
	return &sf
}

// sortDirectionPtr returns a pointer to the given SortDirection value
func sortDirectionPtr(sd SortDirection) *SortDirection {
	return &sd
}
