package duckron

import "github.com/spf13/viper"

type config struct {
	Database DatabaseConfig `yaml:"database"`
	Alerts   AlertsConfig   `yaml:"alerts"`
}

type DatabaseConfig struct {
	Type      string          `yaml:"type"`
	Path      string          `yaml:"path"`
	Snapshot  SnapshotConfig  `yaml:"snapshot"`
	Retention RetentionConfig `yaml:"retention"`
}

type AlertsConfig struct {
	Threshold Threshold `yaml:"threshold"`
}

type Threshold struct {
	Ram  float64 `yaml:"ram"`
	Cpu  float64 `yaml:"cpu"`
	Disk float64 `yaml:"disk"`
}

type SnapshotConfig struct {
	IntervalHours string `yaml:"intervalHours"`
	Destination   string `yaml:"destination"`
	Format        string `yaml:"format"`
}

type RetentionConfig struct {
	IntervalHours string `yaml:"intervalHours"`
}

type ConfigReader struct {
	Reader *viper.Viper
}

func NewConfigReader() *ConfigReader {
	reader := viper.New()
	reader.SetConfigName("duckron")
	reader.AddConfigPath(".")
	reader.SetConfigType("yaml")
	return &ConfigReader{Reader: reader}
}

func (c *ConfigReader) Read() (*config, error) {
	config := &config{}

	err := c.Reader.ReadInConfig()
	if err != nil {
		return nil, err
	}

	if err := c.Reader.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
