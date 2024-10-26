package duckron

import (
	"fmt"
	"os"
	"time"
)

var (
	// ErrConnectionFailed, failed to connect database
	ErrConnectionFailed = newError("connection", "failed to connect database")
	// ErrFolderCreationFailed, failed to create folder
	ErrFolderCreationFailed = newError("folder", "failed to create folder")
	// ErrSnapshotFailed, failed to take snapshot
	ErrSnapshotFailed = newError("snapshot", "failed to take snapshot")
)

type snapshotManager struct {
	client  DatabaseConnection
	options *snapshotOptions
}
type snapshotOptions struct {
	interval    int
	format      string
	destination string
}

func NewSnapshotManager(client DatabaseConnection, options *snapshotOptions) (*snapshotManager, *Error) {
	if options == nil {
		options = &snapshotOptions{
			interval:    60,            // default interval in seconds
			format:      "parquet",     // default format
			destination: "./snapshots", // default destination
		}
	}

	if err := client.Ping(); err != nil {
		return nil, ErrConnectionFailed.wrap(err)
	}

	return &snapshotManager{client: client, options: options}, nil
}

func (sm *snapshotManager) take() *Error {
	if err := createDirectoryIfNotExists(sm.options.destination); err != nil {
		return ErrFolderCreationFailed.wrap(err)
	}

	timer := NewTimer(sm.options.interval)
	timer.Start(
		func() *Error {
			dest := buildSnapshotDestinationPath(sm.options.destination)
			if err := sm.client.Snapshot(sm.options.format, dest); err != nil {
				return ErrSnapshotFailed.wrap(err)
			}
			return nil
		},
	)

	return nil
}

func createDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func buildSnapshotDestinationPath(destination string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/%d", destination, timestamp)
}
