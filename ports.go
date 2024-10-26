package duckron

type DatabaseConnection interface {
	Ping() error
	Close() error
	Snapshot(format string, destination string) error
}
