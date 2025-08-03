package db

import (
	"time"

	"syreclabs.com/go/faker"
)

func (r *eventsRepo) seed() error {
	statement, err := r.db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, name TEXT, advertised_start_time DATETIME, sport_type TEXT, venue TEXT, visible INTEGER)`)
	if err == nil {
		_, err = statement.Exec()
	}

	// Sample sport types and venues
	sportTypes := []string{"football", "basketball", "tennis", "soccer", "baseball", "hockey"}
	venues := []string{"Stadium A", "Arena B", "Court C", "Field D", "Dome E"}

	for i := 1; i <= 100; i++ {
		statement, err = r.db.Prepare(`INSERT OR IGNORE INTO events(id, name, advertised_start_time, sport_type, venue, visible) VALUES (?,?,?,?,?,?)`)
		if err == nil {
			sportIndex := i % len(sportTypes)
			venueIndex := i % len(venues)
			_, err = statement.Exec(
				i,
				faker.Team().Name()+" vs "+faker.Team().Name(), // Create match-style names
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
				sportTypes[sportIndex],
				venues[venueIndex],
				i%2, // Alternate between visible/not visible
			)
		}
	}

	return err
}