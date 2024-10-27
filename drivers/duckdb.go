package drivers

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb"
)

type DuckDBClient struct {
	db *sql.DB
}

func NewDuckDBClient(dataSourceName string) (*DuckDBClient, error) {
	dataSourceName += "?access_mode=read_only&threads=4"
	db, err := sql.Open("duckdb", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DuckDBClient{db: db}, nil
}

func (client *DuckDBClient) Close() error {
	return client.db.Close()
}

func (client *DuckDBClient) exec(query string, args ...interface{}) (sql.Result, error) {
	return client.db.Exec(query, args...)
}

func (client *DuckDBClient) Ping() error {
	return client.db.Ping()
}

func (client *DuckDBClient) Snapshot(format string, destination string) error {
	query := buildSnapshotQuery(format, destination)
	res, err := client.exec(query)
	if err != nil {
		return err
	}
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}
	log.Println("Snapshot created")
	return nil
}

func buildSnapshotQuery(format string, destination string) string {
	query := fmt.Sprintf("EXPORT DATABASE '%s' (FORMAT %s)", destination, format)
	return query
}
