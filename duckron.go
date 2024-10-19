package duckron

import (
	"fmt"
	"os"
	"time"
)

type Duckron struct {
	timer  *Timer
	client *DuckDBClient
	config *config
}

func NewDuckron(config *config) (*Duckron, error) {
	timer := NewTimer(config.Interval)
	client, err := NewDuckDBClient(config.Path)
	if err != nil {
		return nil, err
	}
	return &Duckron{timer: timer, client: client, config: config}, nil
}

func (d *Duckron) RunSnapshotJob() error {
	if err := createDirectoryIfNotExists(d.config.DestinationPath); err != nil {
		return err
	}

	d.timer.Start(
		func() error {
			dest := buildSnapshotDestinationPath(d.config.DestinationPath)
			if err := d.client.Snapshot(d.config.SnapshotFormat, dest); err != nil {
				return err
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
