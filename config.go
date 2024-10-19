package duckron

import "github.com/spf13/viper"

type config struct {
	Path            string
	Interval        int
	DestinationPath string
	SnapshotFormat  string
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
