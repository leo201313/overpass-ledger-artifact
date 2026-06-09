package opt

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

func loadConfig(filename string, config interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open config file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("cannot parse config file: %v", err)
	}

	return nil
}
