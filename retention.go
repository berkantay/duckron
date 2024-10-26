package duckron

type RetentionManager struct {
	timer  *Timer
	client *DuckDBClient
	config *config
}
