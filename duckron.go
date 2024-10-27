package duckron

import (
	"sync"
	"time"

	"github.com/berkantay/duckron/drivers"
)

var (
	// ErrSnapshotManagerFailed is returned when a snapshot manager fails to be created
	ErrSnapshotManagerFailed = newError("snapshot", "failed to create snapshot manager")
	// ErrRetentionManagerFailed is returned when a retention manager fails to be created
	ErrRetentionManagerFailed = newError("retention", "failed to create retention manager")
)

type Duckron struct {
	client   *drivers.DuckDBClient
	config   *config
	services *Services
}

type Services struct {
	snapshotManager  *snapshotManager
	retentionManager *retentionManager
}

func NewDuckron(config *config) (*Duckron, error) {
	client, err := drivers.NewDuckDBClient(config.Database.Path)
	if err != nil {
		return nil, err
	}

	services := &Services{}
	duckron := &Duckron{client: client, config: config, services: services}

	if duckron.isSnapshotConfigured() {
		interval, err := time.ParseDuration(config.Database.Snapshot.IntervalHours)
		if err != nil {
			return nil, err
		}

		snapshotOptions := &snapshotOptions{
			interval:    interval,
			format:      config.Database.Snapshot.Format,
			destination: config.Database.Snapshot.Destination,
		}

		snapshotManager, cerr := NewSnapshotManager(client, snapshotOptions)
		if err != nil {
			return nil, *cerr.unwrap()
		}
		services.snapshotManager = snapshotManager
	}

	if duckron.isRetentionConfigured() {
		interval, err := time.ParseDuration(config.Database.Retention.IntervalHours)

		retentionOptions := &retentionOptions{
			interval: interval,
			path:     config.Database.Snapshot.Destination,
		}

		retentionManager, cerr := NewRetentionManager(client, retentionOptions)
		if err != nil {
			return nil, *cerr.unwrap()
		}
		services.retentionManager = retentionManager
	}

	return duckron, nil
}

func (d *Duckron) Start() *Error {
	var wg sync.WaitGroup
	errChan := make(chan *Error)

	if d.services != nil {
		if d.services.snapshotManager != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := d.services.snapshotManager.take(errChan); err != nil {
					errChan <- err
				}
			}()
		}

		if d.services.retentionManager != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := d.services.retentionManager.clean(errChan); err != nil {
					errChan <- err
				}
			}()
		}
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}

func (d *Duckron) isSnapshotConfigured() bool {
	return d.config.Database.Snapshot != SnapshotConfig{}
}

func (d *Duckron) isRetentionConfigured() bool {
	return d.config.Database.Retention != RetentionConfig{}
}
