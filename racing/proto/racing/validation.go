package racing

import (
	"fmt"
)

const (
	// MaxMeetingIDs defines the maximum number of meeting IDs allowed in a single request
	MaxMeetingIDs = 100
	// MaxMeetingID defines the maximum value for a single meeting ID
	MaxMeetingID = 999999
)

// Validate validates the entire request
func (r *ListRacesRequest) Validate() error {
	if r.Filter != nil {
		return r.Filter.Validate()
	}
	return nil
}

// Validate validates the filter parameters
func (f *ListRacesRequestFilter) Validate() error {
	if err := f.validateMeetingIds(); err != nil {
		return fmt.Errorf("meeting_ids validation failed: %w", err)
	}

	if err := f.validateVisibleOnly(); err != nil {
		return fmt.Errorf("visible_only validation failed: %w", err)
	}

	return nil
}

// validateMeetingIds validates meeting IDs constraints
func (f *ListRacesRequestFilter) validateMeetingIds() error {
	if len(f.MeetingIds) > MaxMeetingIDs {
		return fmt.Errorf("too many meeting IDs: got %d, max allowed %d",
			len(f.MeetingIds), MaxMeetingIDs)
	}

	seen := make(map[int64]bool)
	for i, id := range f.MeetingIds {
		if id <= 0 {
			return fmt.Errorf("invalid meeting ID at position %d: %d (must be positive)", i, id)
		}

		if id > MaxMeetingID {
			return fmt.Errorf("meeting ID too large at position %d: %d (max: %d)",
				i, id, MaxMeetingID)
		}

		if seen[id] {
			return fmt.Errorf("duplicate meeting ID: %d", id)
		}
		seen[id] = true
	}

	return nil
}

// validateVisibleOnly validates visible_only field (future extensibility)
func (f *ListRacesRequestFilter) validateVisibleOnly() error {
	return nil
}
