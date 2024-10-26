package duckron

var (
	// ErrSnapshotManagerFailed is returned when a snapshot manager fails to be created
	ErrSnapshotManagerFailed = newError("snapshot", "failed to create snapshot manager")
)

type Duckron struct {
	client *DuckDBClient
	config *config
}

type Services struct {
	snapshotManager *snapshotManager
}

func NewDuckron(config *config) (*Duckron, error) {
	client, err := NewDuckDBClient(config.Database.Path)
	if err != nil {
		return nil, err
	}
	return &Duckron{client: client, config: config}, nil
}

func (d *Duckron) Start() *Error {
	snapshotOptions := &snapshotOptions{
		interval:    d.config.Database.Snapshot.Interval,
		format:      d.config.Database.Snapshot.Format,
		destination: d.config.Database.Snapshot.Destination,
	}

	if d.isSnapshotConfigured() {
		snapshotManager, err := NewSnapshotManager(d.client, snapshotOptions)
		if err != nil {
			return ErrSnapshotManagerFailed.wrap(*err.rootErr)
		}
		services := &Services{snapshotManager: snapshotManager}
		return services.snapshotManager.take()
	} else {
		return nil
	}
}

func (d *Duckron) isSnapshotConfigured() bool {
	return d.config.Database.Snapshot != SnapshotConfig{}
}
