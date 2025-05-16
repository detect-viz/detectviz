package eventlog

// EventStore defines the interface for storing and retrieving event logs.
type EventStore interface {
	SaveEvent(e *EventLog) error
	ListRecent(limit int) ([]EventLog, error)
}
