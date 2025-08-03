package db

import (
	"database/sql"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("setupTestDB() failed to open database: %v", err)
	}

	query := `
		CREATE TABLE races (
			id INTEGER PRIMARY KEY,
			meeting_id INTEGER,
			name TEXT,
			number INTEGER,
			visible INTEGER,
			advertised_start_time DATETIME
		)
	`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("setupTestDB() failed to create table: %v", err)
	}

	return db
}

// insertTestRace inserts a test race into the database
func insertTestRace(t *testing.T, db *sql.DB, id, meetingID, number int, name string, visible bool, startTime time.Time) {
	t.Helper()

	visibleInt := 0
	if visible {
		visibleInt = 1
	}

	query := `
		INSERT INTO races (id, meeting_id, name, number, visible, advertised_start_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := db.Exec(query, id, meetingID, name, number, visibleInt, startTime.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insertTestRace(id=%d) failed: %v", id, err)
	}
}

// boolPtr returns a pointer to the given bool value
func boolPtr(b bool) *bool {
	return &b
}

// sortFieldPtr returns a pointer to the given SortField value
func sortFieldPtr(sf racing.SortField) *racing.SortField {
	return &sf
}

// sortDirectionPtr returns a pointer to the given SortDirection value
func sortDirectionPtr(sd racing.SortDirection) *racing.SortDirection {
	return &sd
}

func TestApplyFilter(t *testing.T) {
	tests := []struct {
		name      string
		filter    *racing.ListRacesRequestFilter
		wantQuery string
		wantArgs  []interface{}
	}{
		{
			name:      "nil filter returns original query",
			filter:    nil,
			wantQuery: "SELECT * FROM races",
			wantArgs:  nil,
		},
		{
			name:      "empty filter returns original query",
			filter:    &racing.ListRacesRequestFilter{},
			wantQuery: "SELECT * FROM races",
			wantArgs:  nil,
		},
		{
			name: "visible only true adds visible clause",
			filter: &racing.ListRacesRequestFilter{
				VisibleOnly: boolPtr(true),
			},
			wantQuery: "SELECT * FROM races WHERE visible = 1",
			wantArgs:  nil,
		},
		{
			name: "visible only false does not add visible clause",
			filter: &racing.ListRacesRequestFilter{
				VisibleOnly: boolPtr(false),
			},
			wantQuery: "SELECT * FROM races",
			wantArgs:  nil,
		},
		{
			name: "meeting ids filter creates IN clause",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds: []int64{1, 2, 3},
			},
			wantQuery: "SELECT * FROM races WHERE meeting_id IN (?,?,?)",
			wantArgs:  []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			name: "combined filters work together",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds:  []int64{1, 2},
				VisibleOnly: boolPtr(true),
			},
			wantQuery: "SELECT * FROM races WHERE meeting_id IN (?,?) AND visible = 1",
			wantArgs:  []interface{}{int64(1), int64(2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &racesRepo{}
			baseQuery := "SELECT * FROM races"

			gotQuery, gotArgs := repo.applyFilter(baseQuery, tt.filter)

			if gotQuery != tt.wantQuery {
				t.Errorf("applyFilter() query = %q, want %q", gotQuery, tt.wantQuery)
			}

			// Use lenient comparison that allows nil and empty slice equivalence
			argsEqual := (tt.wantArgs == nil && gotArgs == nil) ||
				(tt.wantArgs == nil && len(gotArgs) == 0) ||
				(len(tt.wantArgs) == 0 && gotArgs == nil) ||
				cmp.Equal(tt.wantArgs, gotArgs)

			if !argsEqual {
				t.Errorf("applyFilter() args = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestRacesRepo_List(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := NewRacesRepo(db)

	// Setup test data
	now := time.Now()
	testRaces := []struct {
		id        int
		meetingID int
		name      string
		number    int
		visible   bool
		startTime time.Time
	}{
		{1, 1, "Visible Race 1", 1, true, now.Add(time.Hour)},
		{2, 1, "Hidden Race 1", 2, false, now.Add(2 * time.Hour)},
		{3, 2, "Visible Race 2", 1, true, now.Add(3 * time.Hour)},
		{4, 2, "Hidden Race 2", 2, false, now.Add(4 * time.Hour)},
	}

	for _, race := range testRaces {
		insertTestRace(t, db, race.id, race.meetingID, race.number, race.name, race.visible, race.startTime)
	}

	tests := []struct {
		name    string
		filter  *racing.ListRacesRequestFilter
		wantIDs []int64
	}{
		{
			name:    "no filter returns all races",
			filter:  &racing.ListRacesRequestFilter{},
			wantIDs: []int64{1, 2, 3, 4},
		},
		{
			name: "visible only true returns only visible races",
			filter: &racing.ListRacesRequestFilter{
				VisibleOnly: boolPtr(true),
			},
			wantIDs: []int64{1, 3},
		},
		{
			name: "visible only false returns all races",
			filter: &racing.ListRacesRequestFilter{
				VisibleOnly: boolPtr(false),
			},
			wantIDs: []int64{1, 2, 3, 4},
		},
		{
			name: "meeting ids filter works correctly",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds: []int64{1},
			},
			wantIDs: []int64{1, 2},
		},
		{
			name: "combined meeting ids and visible only filters",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds:  []int64{1},
				VisibleOnly: boolPtr(true),
			},
			wantIDs: []int64{1},
		},
		{
			name: "non existent meeting id returns empty results",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds: []int64{999},
			},
			wantIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRaces, err := repo.List(tt.filter)
			if err != nil {
				t.Fatalf("List(%+v) failed: %v", tt.filter, err)
			}

			var gotIDs []int64
			for _, race := range gotRaces {
				gotIDs = append(gotIDs, race.Id)
			}

			if tt.wantIDs == nil && len(gotIDs) == 0 {
				// Test passes: expected nil, got empty slice, they are equivalent
				return
			}

			sortOpt := cmpopts.SortSlices(func(a, b int64) bool { return a < b })
			if diff := cmp.Diff(tt.wantIDs, gotIDs, sortOpt); diff != "" {
				t.Errorf("List(%+v) race IDs mismatch (-want +got):\n%s", tt.filter, diff)
			}

			// Additional validations for each race
			for _, race := range gotRaces {
				// Validate required fields are not zero values
				if race.Id <= 0 {
					t.Errorf("List(%+v): race.Id = %d, want > 0", tt.filter, race.Id)
				}
				if race.MeetingId <= 0 {
					t.Errorf("List(%+v): race.MeetingId = %d, want > 0", tt.filter, race.MeetingId)
				}
				if race.Name == "" {
					t.Errorf("List(%+v): race.Name is empty for race ID %d", tt.filter, race.Id)
				}
				if race.Number <= 0 {
					t.Errorf("List(%+v): race.Number = %d, want > 0 for race ID %d", tt.filter, race.Number, race.Id)
				}
				if race.AdvertisedStartTime == nil {
					t.Errorf("List(%+v): race.AdvertisedStartTime is nil for race ID %d", tt.filter, race.Id)
				}

				// Validate status field is properly set
				gotTime, err := ptypes.Timestamp(race.AdvertisedStartTime)
				if err == nil {
					expectedStatus := racing.RaceStatus_OPEN
					if gotTime.Before(time.Now()) {
						expectedStatus = racing.RaceStatus_CLOSED
					}
					if race.Status != expectedStatus {
						t.Errorf("List(%+v): race ID %d has status %v, want %v based on start time %v",
							tt.filter, race.Id, race.Status, expectedStatus, gotTime)
					}
				}

				if tt.filter != nil && tt.filter.VisibleOnly != nil && *tt.filter.VisibleOnly {
					if !race.Visible {
						t.Errorf("List(%+v): expected only visible races, but race ID %d has visible = %t",
							tt.filter, race.Id, race.Visible)
					}
				}

				if tt.filter != nil && len(tt.filter.MeetingIds) > 0 {
					found := false
					for _, meetingID := range tt.filter.MeetingIds {
						if race.MeetingId == meetingID {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("List(%+v): race ID %d has meeting_id %d, which is not in filter",
							tt.filter, race.Id, race.MeetingId)
					}
				}
			}
		})
	}
}

func TestRacesRepo_List_DataIntegrity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRacesRepo(db)

	// Insert test race with specific known values
	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	insertTestRace(t, db, 1, 123, 5, "Test Race", true, testTime)

	gotRaces, err := repo.List(&racing.ListRacesRequestFilter{
		VisibleOnly: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(gotRaces) != 1 {
		t.Fatalf("List() returned %d races, want 1", len(gotRaces))
	}

	race := gotRaces[0]

	// Test each field individually for better error messages
	if race.Id != 1 {
		t.Errorf("List() race.Id = %d, want 1", race.Id)
	}
	if race.MeetingId != 123 {
		t.Errorf("List() race.MeetingId = %d, want 123", race.MeetingId)
	}
	if race.Name != "Test Race" {
		t.Errorf("List() race.Name = %q, want %q", race.Name, "Test Race")
	}
	if race.Number != 5 {
		t.Errorf("List() race.Number = %d, want 5", race.Number)
	}
	if !race.Visible {
		t.Errorf("List() race.Visible = %t, want true", race.Visible)
	}

	// Check timestamp conversion
	if race.AdvertisedStartTime == nil {
		t.Errorf("List() race.AdvertisedStartTime is nil")
	} else {
		gotTime, err := ptypes.Timestamp(race.AdvertisedStartTime)
		if err != nil {
			t.Errorf("List() failed to convert timestamp: %v", err)
		} else if !testTime.Equal(gotTime) {
			t.Errorf("List() race.AdvertisedStartTime = %v, want %v", gotTime, testTime)
		}
	}
}

func TestRacesRepo_List_DatabaseErrors(t *testing.T) {
	// Test with closed database to simulate database errors
	db := setupTestDB(t)
	db.Close() // Close immediately to cause errors

	repo := NewRacesRepo(db)

	_, err := repo.List(&racing.ListRacesRequestFilter{})
	if err == nil {
		t.Error("List() with closed database returned no error, want error")
	}
}

func TestNewRacesRepo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRacesRepo(db)
	if repo == nil {
		t.Error("NewRacesRepo() returned nil, want non-nil repo")
	}
}

func TestApplySorting(t *testing.T) {
	tests := []struct {
		name      string
		filter    *racing.ListRacesRequestFilter
		baseQuery string
		want      string
	}{
		{
			name:      "nil filter uses default sorting",
			filter:    nil,
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY advertised_start_time ASC",
		},
		{
			name:      "empty filter uses default sorting",
			filter:    &racing.ListRacesRequestFilter{},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY advertised_start_time ASC",
		},
		{
			name: "sort by name ascending",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NAME),
				SortDirection: sortDirectionPtr(racing.SortDirection_ASC),
			},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY name ASC",
		},
		{
			name: "sort by name descending",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NAME),
				SortDirection: sortDirectionPtr(racing.SortDirection_DESC),
			},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY name DESC",
		},
		{
			name: "sort by number ascending",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NUMBER),
				SortDirection: sortDirectionPtr(racing.SortDirection_ASC),
			},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY number ASC",
		},
		{
			name: "sort by advertised start time descending",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_ADVERTISED_START_TIME),
				SortDirection: sortDirectionPtr(racing.SortDirection_DESC),
			},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY advertised_start_time DESC",
		},
		{
			name: "only sort field specified defaults to ASC",
			filter: &racing.ListRacesRequestFilter{
				SortField: sortFieldPtr(racing.SortField_NUMBER),
			},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY number ASC",
		},
		{
			name: "only sort direction specified uses default field",
			filter: &racing.ListRacesRequestFilter{
				SortDirection: sortDirectionPtr(racing.SortDirection_DESC),
			},
			baseQuery: "SELECT * FROM races",
			want:      "SELECT * FROM races ORDER BY advertised_start_time DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &racesRepo{}
			got := repo.applySorting(tt.baseQuery, tt.filter)

			if got != tt.want {
				t.Errorf("applySorting() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRacesRepo_List_Sorting(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := NewRacesRepo(db)

	// Setup test data with different start times for sorting
	now := time.Now()
	testRaces := []struct {
		id        int
		meetingID int
		name      string
		number    int
		visible   bool
		startTime time.Time
	}{
		{1, 1, "Charlie Race", 3, true, now.Add(3 * time.Hour)}, // Latest time
		{2, 1, "Alpha Race", 1, true, now.Add(1 * time.Hour)},   // Earliest time
		{3, 1, "Bravo Race", 2, true, now.Add(2 * time.Hour)},   // Middle time
	}

	for _, race := range testRaces {
		insertTestRace(t, db, race.id, race.meetingID, race.number, race.name, race.visible, race.startTime)
	}

	tests := []struct {
		name      string
		filter    *racing.ListRacesRequestFilter
		wantOrder []int64 // Expected race IDs in order
	}{
		{
			name:      "default sorting by advertised_start_time ASC",
			filter:    &racing.ListRacesRequestFilter{},
			wantOrder: []int64{2, 3, 1}, // Earliest to latest
		},
		{
			name: "sort by advertised_start_time DESC",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_ADVERTISED_START_TIME),
				SortDirection: sortDirectionPtr(racing.SortDirection_DESC),
			},
			wantOrder: []int64{1, 3, 2}, // Latest to earliest
		},
		{
			name: "sort by name ASC",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NAME),
				SortDirection: sortDirectionPtr(racing.SortDirection_ASC),
			},
			wantOrder: []int64{2, 3, 1}, // Alpha, Bravo, Charlie
		},
		{
			name: "sort by name DESC",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NAME),
				SortDirection: sortDirectionPtr(racing.SortDirection_DESC),
			},
			wantOrder: []int64{1, 3, 2}, // Charlie, Bravo, Alpha
		},
		{
			name: "sort by number ASC",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NUMBER),
				SortDirection: sortDirectionPtr(racing.SortDirection_ASC),
			},
			wantOrder: []int64{2, 3, 1}, // Numbers 1, 2, 3
		},
		{
			name: "sort by number DESC",
			filter: &racing.ListRacesRequestFilter{
				SortField:     sortFieldPtr(racing.SortField_NUMBER),
				SortDirection: sortDirectionPtr(racing.SortDirection_DESC),
			},
			wantOrder: []int64{1, 3, 2}, // Numbers 3, 2, 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRaces, err := repo.List(tt.filter)
			if err != nil {
				t.Fatalf("List(%+v) failed: %v", tt.filter, err)
			}

			if len(gotRaces) != len(tt.wantOrder) {
				t.Fatalf("List(%+v) returned %d races, want %d", tt.filter, len(gotRaces), len(tt.wantOrder))
			}

			var gotOrder []int64
			for _, race := range gotRaces {
				gotOrder = append(gotOrder, race.Id)
			}

			if diff := cmp.Diff(tt.wantOrder, gotOrder); diff != "" {
				t.Errorf("List(%+v) race order mismatch (-want +got):\n%s", tt.filter, diff)
			}
		})
	}
}

func TestRacesRepo_List_StatusLogic(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	}()

	repo := NewRacesRepo(db)

	// Setup test data with past and future times
	now := time.Now()
	testRaces := []struct {
		id             int
		name           string
		startTime      time.Time
		expectedStatus racing.RaceStatus
	}{
		{1, "Past Race", now.Add(-1 * time.Hour), racing.RaceStatus_CLOSED},
		{2, "Future Race", now.Add(1 * time.Hour), racing.RaceStatus_OPEN},
		{3, "Very Past Race", now.Add(-24 * time.Hour), racing.RaceStatus_CLOSED},
		{4, "Very Future Race", now.Add(24 * time.Hour), racing.RaceStatus_OPEN},
	}

	for _, race := range testRaces {
		insertTestRace(t, db, race.id, 1, 1, race.name, true, race.startTime)
	}

	gotRaces, err := repo.List(&racing.ListRacesRequestFilter{})
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(gotRaces) != len(testRaces) {
		t.Fatalf("List() returned %d races, want %d", len(gotRaces), len(testRaces))
	}

	// Create a map for easier lookup
	raceMap := make(map[int64]*racing.Race)
	for _, race := range gotRaces {
		raceMap[race.Id] = race
	}

	for _, expectedRace := range testRaces {
		gotRace, exists := raceMap[int64(expectedRace.id)]
		if !exists {
			t.Errorf("Expected race ID %d not found in results", expectedRace.id)
			continue
		}

		if gotRace.Status != expectedRace.expectedStatus {
			t.Errorf("Race ID %d (%s) has status %v, want %v",
				expectedRace.id, expectedRace.name, gotRace.Status, expectedRace.expectedStatus)
		}
	}
}
