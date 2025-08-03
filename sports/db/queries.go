package db

const (
	eventsList    = "list"
	eventsGetByID = "getByID"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventsList: `
			SELECT 
				id, 
				name, 
				advertised_start_time,
				sport_type,
				venue,
				visible
			FROM events
		`,
		eventsGetByID: `
			SELECT 
				id, 
				name, 
				advertised_start_time,
				sport_type,
				venue,
				visible
			FROM events 
			WHERE id = ?
		`,
	}
}
