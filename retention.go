package duckron

import (
	"io/fs"
	"log"
	"os"
	"time"
)

var (
	ErrFolderDeletionFailed = newError("folder", "failed to delete snapshot")
)

type folderMetadata struct {
	name      string
	createdAt time.Time
}

type retentionManager struct {
	client  DatabaseConnection
	options *retentionOptions
	timer   *Timer
}

type retentionOptions struct {
	interval time.Duration
	path     string
}

func NewRetentionManager(client DatabaseConnection, options *retentionOptions) (*retentionManager, *Error) {
	if options.interval == 0 {
		options = &retentionOptions{
			interval: time.Hour * 24,
		}
	}

	if err := client.Ping(); err != nil {
		return nil, ErrConnectionFailed.wrap(err)
	}

	timer := NewTimer(options.interval)

	return &retentionManager{client: client, options: options, timer: timer}, nil
}

func (rm *retentionManager) clean(errChan chan *Error) *Error {
	go func(errChan chan *Error) {

		rm.timer.Start(
			func() *Error {
				log.Println("Cleaning snapshots")
				files, err := rm.readFoldersInPath(rm.options.path)
				if err != nil {
					return ErrSnapshotFailed.wrap(err)
				}

				fMeta, err := rm.readSnapshotTimestamps(files)
				if err != nil {
					return ErrSnapshotFailed.wrap(err)
				}

				if err := rm.deleteOldSnapshots(fMeta); err != nil {
					errChan <- err
					return ErrFolderDeletionFailed
				}
				return nil
			},
		)
	}(errChan)

	return nil
}

func (rm *retentionManager) readFoldersInPath(path string) ([]fs.FileInfo, error) {
	folder, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer folder.Close()

	files, err := folder.Readdir(0)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (rm *retentionManager) readSnapshotTimestamps(file []fs.FileInfo) ([]*folderMetadata, error) {
	var fMeta []*folderMetadata
	if len(file) == 0 {
		return fMeta, nil
	}

	for _, file := range file {
		fMeta = append(fMeta, &folderMetadata{
			name:      file.Name(),
			createdAt: file.ModTime(),
		})
	}
	return fMeta, nil
}

func (rm *retentionManager) deleteOldSnapshots(files []*folderMetadata) *Error {
	for _, file := range files {
		if time.Since(file.createdAt).Hours() > float64(rm.options.interval.Hours()) {
			log.Println("Deleting snapshot", file.name)
			log.Println("Snapshot created at", file.createdAt)
			log.Println("Snapshot age", time.Since(file.createdAt).Hours())

			err := os.RemoveAll(rm.options.path + "/" + file.name)
			if err != nil {
				return ErrFolderDeletionFailed.wrap(err)
			}
			log.Println("Snapshot deleted", file.name)
		}
	}
	return nil
}
