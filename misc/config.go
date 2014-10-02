package misc

import (
	"github.com/mitchellh/mapstructure"
	"github.com/srhnsn/go-utils/log"
	"gopkg.in/yaml.v2"
)

func LoadConfig(filename string, f func(name string) ([]byte, error), configType interface{}) {
	rawConfig := getRawConfig(filename, f)

	err := mapstructure.Decode(rawConfig, configType)

	if err == nil {
		log.Trace.Printf("Loaded configuration from %s", filename)
	} else {
		log.Error.Fatalf("Cannot load config: %s", err)
	}
}

func getRawConfig(filename string, f func(name string) ([]byte, error)) map[string]interface{} {
	data, err := f(filename)

	if err != nil {
		log.Error.Fatalf("Cannot read configuration file: %s", err)
	}

	config := make(map[string]interface{})
	yaml.Unmarshal(data, config)
	return config
}
