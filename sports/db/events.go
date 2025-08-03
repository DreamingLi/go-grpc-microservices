package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// EventsRepo provides repository access to sports events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// List will return a list of events.
	List(filter *sports.ListEventsRequestFilter) ([]*sports.Event, error)

	// GetByID will return a single event by its ID.
	GetByID(id int64) (*sports.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository dummy data.
func (r *eventsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = r.seed()
	})

	return err
}

// List retrieves events from the database based on the provided filter.
// It supports filtering by sport types and visibility status.
// Results are ordered by advertised_start_time ASC by default, or by the specified sort field and direction.
func (r *eventsRepo) List(filter *sports.ListEventsRequestFilter) ([]*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsList]

	query, args = r.applyFilter(query, filter)
	query = r.applySorting(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanEvents(rows)
}

// GetByID retrieves a single event from the database by its ID.
// Returns the event if found, or an error if not found or database error occurs.
func (r *eventsRepo) GetByID(id int64) (*sports.Event, error) {
	query := getEventQueries()[eventsGetByID]

	row := r.db.QueryRow(query, id)

	var event sports.Event
	var advertisedStart time.Time

	err := row.Scan(&event.Id, &event.Name, &advertisedStart, &event.SportType, &event.Venue, &event.Visible)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event with ID %d not found", id)
		}
		return nil, err
	}

	ts, err := ptypes.TimestampProto(advertisedStart)
	if err != nil {
		return nil, err
	}

	event.AdvertisedStartTime = ts

	// Set event status based on advertised start time
	setEventStatus(&event, advertisedStart)

	return &event, nil
}

// applyFilter modifies the base query to include WHERE clauses based on the filter.
// It returns the modified query string and the corresponding arguments for parameterized queries.
func (r *eventsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.SportTypes) > 0 {
		placeholders := strings.Repeat("?,", len(filter.SportTypes)-1) + "?"
		clauses = append(clauses, "sport_type IN ("+placeholders+")")

		for _, sportType := range filter.SportTypes {
			args = append(args, sportType)
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
func (r *eventsRepo) applySorting(query string, filter *sports.ListEventsRequestFilter) string {
	var sortField string
	var sortDirection string

	// Determine sort field (default to advertised_start_time)
	if filter != nil && filter.SortField != nil {
		switch *filter.SortField {
		case sports.SortField_NAME:
			sortField = "name"
		case sports.SortField_SPORT_TYPE:
			sortField = "sport_type"
		case sports.SortField_ADVERTISED_START_TIME:
			sortField = "advertised_start_time"
		default:
			sortField = "advertised_start_time"
		}
	} else {
		sortField = "advertised_start_time"
	}

	// Determine sort direction (default to ASC)
	if filter != nil && filter.SortDirection != nil && *filter.SortDirection == sports.SortDirection_DESC {
		sortDirection = "DESC"
	} else {
		sortDirection = "ASC"
	}

	return query + " ORDER BY " + sortField + " " + sortDirection
}

func (r *eventsRepo) scanEvents(
	rows *sql.Rows,
) ([]*sports.Event, error) {
	var events []*sports.Event

	for rows.Next() {
		var event sports.Event
		var advertisedStart time.Time

		if err := rows.Scan(&event.Id, &event.Name, &advertisedStart, &event.SportType, &event.Venue, &event.Visible); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts

		// Set event status based on advertised start time
		setEventStatus(&event, advertisedStart)

		events = append(events, &event)
	}

	return events, nil
}

// setEventStatus sets the event status based on the advertised start time.
// Events with advertised start time in the past are marked as CLOSED, others as OPEN.
func setEventStatus(event *sports.Event, advertisedStart time.Time) {
	event.Status = sports.EventStatus_OPEN
	if advertisedStart.Before(time.Now()) {
		event.Status = sports.EventStatus_CLOSED
	}
}
