package duckron

import (
	"fmt"
	"log"
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
	timer   *Timer
}
type snapshotOptions struct {
	interval    time.Duration
	format      string
	destination string
}

func NewSnapshotManager(client DatabaseConnection, options *snapshotOptions) (*snapshotManager, *Error) {
	if options == nil {
		options = &snapshotOptions{
			interval:    60,
			format:      "parquet",
			destination: "./snapshots",
		}
	}

	if err := client.Ping(); err != nil {
		return nil, ErrConnectionFailed.wrap(err)
	}
	timer := NewTimer(options.interval)

	return &snapshotManager{client: client, options: options, timer: timer}, nil
}

func (sm *snapshotManager) take(errChan chan *Error) *Error {
	if err := createDirectoryIfNotExists(sm.options.destination); err != nil {
		return ErrFolderCreationFailed.wrap(err)
	}

	go func(errChan chan *Error) {
		sm.timer.Start(
			func() *Error {
				log.Println("Taking snapshot")
				dest := buildSnapshotDestinationPath(sm.options.destination)
				if err := sm.client.Snapshot(sm.options.format, dest); err != nil {
					errChan <- ErrSnapshotFailed.wrap(err)
					return ErrSnapshotFailed.wrap(err)
				}
				return nil
			},
		)
	}(errChan)

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
