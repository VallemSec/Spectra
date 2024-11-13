package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"main/types"
	"os"
)

func getAndUnmarshalConfigFile(configFileName string) (types.ConfigFile, error) {
	yamlFile, err := os.ReadFile(configFileName)
	if err != nil {
		return types.ConfigFile{}, fmt.Errorf("could not read %s read config error: %v", configFileName, err)
	}

	var config types.ConfigFile
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return types.ConfigFile{}, fmt.Errorf("failed to unmarshall the config, this is typically due to a malformed config Unmarshalling error: %v", err)
	}

	return config, nil
}
