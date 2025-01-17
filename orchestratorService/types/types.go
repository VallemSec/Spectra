package types

import (
	"encoding/json"
	"fmt"
)

type JSONbody struct {
	Target string `json:"target"`
}

type ConfigFile struct {
	DiscoveryRunners []string                `yaml:"discovery_runners"`
	AlwaysRun        []string                `yaml:"always_run"`
	Runners          map[string]RunnerConfig `yaml:"runners"`
}

type RunnerConfig struct {
	CmdArgs       []string            `yaml:"cmdargs"`
	Report        bool                `yaml:"report"`
	ContainerName string              `yaml:"container_name"`
	Image         string              `yaml:"image"`
	ImageVersion  string              `yaml:"image_version"`
	ParserPlugin  string              `yaml:"parser_plugin"`
	DecodyRule    []string            `yaml:"decody_rule,omitempty"`
	Results       map[string][]string `yaml:"results"`
	Tty           bool                `yaml:"tty"`
}

type DockerRunnerResult struct {
	Stdout []string `json:"stdout"`
	Stderr []string `json:"stderr"`
}

type ParserOutputJson struct {
	ScannerName string   `json:"name"`
	Results     []Result `json:"results"`
}

type Result struct {
	Short   string   `json:"short"`
	Long    string   `json:"long"`
	PassRes []string `json:"pass_results,omitempty"`
}

// UnmarshalJSON custom unmarshaller for Result
func (r *Result) UnmarshalJSON(data []byte) error {
	type Alias Result
	aux := &struct {
		PassResults interface{} `json:"pass_results"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.PassResults.(type) {
	case string:
		r.PassRes = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				r.PassRes = append(r.PassRes, str)
			} else {
				return fmt.Errorf("unexpected type for pass_results field")
			}
		}
	case nil:
		r.PassRes = []string{}
	default:
		return fmt.Errorf("unexpected type for pass_results field")
	}

	return nil
}

type RunnerJSON struct {
	ContainerName    string   `json:"containerName"`
	ContainerTag     string   `json:"containerTag"`
	ContainerCommand []string `json:"containerCommand"`
	Volumes          []string `json:"volume"`
	Networks         []string `json:"network"`
	Env              []string `json:"env"`
	Tty              bool     `json:"tty"`
}

type DecodyInput struct {
	Name    string   `json:"scanner_name"`
	Rules   []string `json:"rules"`
	Results []Result `json:"results"`
}

type EmailReturn struct {
	Service string `json:"service"`
	Date    string `json:"date"`
	Icon    string `json:"icon"`
}
