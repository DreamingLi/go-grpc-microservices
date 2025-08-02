package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

// List retrieves races from the database based on the provided filter.
// It supports filtering by meeting IDs and visibility status.
// Results are ordered by advertised_start_time ASC by default, or by the specified sort field and direction.
func (r *racesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)
	query = r.applySorting(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

// applyFilter modifies the base query to include WHERE clauses based on the filter.
// It returns the modified query string and the corresponding arguments for parameterized queries.
func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		placeholders := strings.Repeat("?,", len(filter.MeetingIds)-1) + "?"
		clauses = append(clauses, "meeting_id IN ("+placeholders+")")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	if filter.VisibleOnly != nil && *filter.VisibleOnly {
		clauses = append(clauses, "visible = 1")
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

// applySorting adds ORDER BY clause to the query based on the filter's sort preferences.
// Defaults to ORDER BY advertised_start_time ASC if no sort field is specified.
func (r *racesRepo) applySorting(query string, filter *racing.ListRacesRequestFilter) string {
	var sortField string
	var sortDirection string

	// Determine sort field (default to advertised_start_time)
	if filter != nil && filter.SortField != nil {
		switch *filter.SortField {
		case racing.SortField_NAME:
			sortField = "name"
		case racing.SortField_NUMBER:
			sortField = "number"
		case racing.SortField_ADVERTISED_START_TIME:
			sortField = "advertised_start_time"
		default:
			sortField = "advertised_start_time"
		}
	} else {
		sortField = "advertised_start_time"
	}

	// Determine sort direction (default to ASC)
	if filter != nil && filter.SortDirection != nil && *filter.SortDirection == racing.SortDirection_DESC {
		sortDirection = "DESC"
	} else {
		sortDirection = "ASC"
	}

	return query + " ORDER BY " + sortField + " " + sortDirection
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
