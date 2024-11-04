package duckron

import (
	"fmt"
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
	client    *drivers.DuckDBClient
	config    *config
	services  *Services
	alertChan chan int
	errChan   chan *Error
}

type Services struct {
	snapshotManager  *snapshotManager
	retentionManager *retentionManager
	alertManager     *alertManager
}

func NewDuckron(config *config) (*Duckron, error) {
	errChan := make(chan *Error)
	alertChan := make(chan int)

	client, err := drivers.NewDuckDBClient(config.Database.Path)
	if err != nil {
		return nil, err
	}

	services := &Services{}
	duckron := &Duckron{client: client, config: config, services: services, alertChan: alertChan, errChan: errChan}

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

	if duckron.isAlertConfigured() {
		alertOptions := &alertOptions{
			ramThreshold:  config.Alerts.Threshold.Ram,
			cpuThreshold:  config.Alerts.Threshold.Cpu,
			diskThreshold: config.Alerts.Threshold.Disk,
		}

		alertManager := NewAlertManager(alertOptions, alertChan, errChan)

		services.alertManager = alertManager
	}

	return duckron, nil
}

func (d *Duckron) Start() *Error {
	var wg sync.WaitGroup

	go func() {
		for err := range d.errChan {
			if err != nil {
				fmt.Println("Error:", err)
			}
		}
	}()

	go func() {
		for alert := range d.alertChan {
			switch alert {
			case CPU_THRESHOLD_EXCEEDED:
				fmt.Println("CPU threshold exceeded")
			case RAM_THRESHOLD_EXCEEDED:
				fmt.Println("RAM threshold exceeded")
			case DISK_THRESHOLD_EXCEEDED:
				fmt.Println("Disk threshold exceeded")
			}
		}
	}()

	if d.services != nil {
		if d.services.snapshotManager != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := d.services.snapshotManager.take(d.errChan); err != nil {
					d.errChan <- err
				}
			}()
		}

		if d.services.retentionManager != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := d.services.retentionManager.clean(d.errChan); err != nil {
					d.errChan <- err
				}
			}()
		}

		if d.services.alertManager != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				d.services.alertManager.monitor()
			}()
		}
	}

	wg.Wait()
	close(d.errChan)

	if len(d.errChan) > 0 {
		return <-d.errChan
	}

	return nil
}

func (d *Duckron) isSnapshotConfigured() bool {
	return d.config.Database.Snapshot != SnapshotConfig{}
}

func (d *Duckron) isRetentionConfigured() bool {
	return d.config.Database.Retention != RetentionConfig{}
}

func (d *Duckron) isAlertConfigured() bool {
	return d.config.Alerts != AlertsConfig{}
}
