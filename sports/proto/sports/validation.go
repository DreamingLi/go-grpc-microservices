package sports

import (
	"fmt"
	"strings"
)

const (
	// MaxSportTypes defines the maximum number of sport types allowed in a single request
	MaxSportTypes = 50
	// MaxSportTypeLength defines the maximum length for a sport type string
	MaxSportTypeLength = 100
)

// Validate validates the GetEvent request
func (r *GetEventRequest) Validate() error {
	if r.Id <= 0 {
		return fmt.Errorf("invalid event ID: %d (must be positive)", r.Id)
	}
	return nil
}

// Validate validates the entire ListEvents request
func (r *ListEventsRequest) Validate() error {
	if r.Filter != nil {
		return r.Filter.Validate()
	}
	return nil
}

// Validate validates the filter parameters
func (f *ListEventsRequestFilter) Validate() error {
	if err := f.validateSportTypes(); err != nil {
		return fmt.Errorf("sport_types validation failed: %w", err)
	}

	if err := f.validateVisibleOnly(); err != nil {
		return fmt.Errorf("visible_only validation failed: %w", err)
	}

	if err := f.validateSorting(); err != nil {
		return fmt.Errorf("sorting validation failed: %w", err)
	}

	return nil
}

// validateSportTypes validates sport types constraints
func (f *ListEventsRequestFilter) validateSportTypes() error {
	if len(f.SportTypes) > MaxSportTypes {
		return fmt.Errorf("too many sport types: got %d, max allowed %d",
			len(f.SportTypes), MaxSportTypes)
	}

	seen := make(map[string]bool)
	for i, sportType := range f.SportTypes {
		sportType = strings.TrimSpace(sportType)
		
		if sportType == "" {
			return fmt.Errorf("empty sport type at position %d", i)
		}

		if len(sportType) > MaxSportTypeLength {
			return fmt.Errorf("sport type too long at position %d: %d characters (max: %d)",
				i, len(sportType), MaxSportTypeLength)
		}

		if seen[sportType] {
			return fmt.Errorf("duplicate sport type: %s", sportType)
		}
		seen[sportType] = true
	}

	return nil
}

// validateVisibleOnly validates visible_only field (future extensibility)
func (f *ListEventsRequestFilter) validateVisibleOnly() error {
	return nil
}

// validateSorting validates sorting parameters
func (f *ListEventsRequestFilter) validateSorting() error {
	// Validate sort field
	if f.SortField != nil {
		switch *f.SortField {
		case SortField_ADVERTISED_START_TIME, SortField_NAME, SortField_SPORT_TYPE:
			// Valid sort fields
		default:
			return fmt.Errorf("invalid sort field: %v", *f.SortField)
		}
	}

	// Validate sort direction
	if f.SortDirection != nil {
		switch *f.SortDirection {
		case SortDirection_ASC, SortDirection_DESC:
			// Valid sort directions
		default:
			return fmt.Errorf("invalid sort direction: %v", *f.SortDirection)
		}
	}

	return nil
}