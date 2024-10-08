package types

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
}

type ParserOutputJson struct {
	ScannerName string   `json:"scanner_name"`
	Results     []Result `json:"results"`
}

type Result struct {
	Short   string   `json:"short"`
	Long    string   `json:"long"`
	PassRes []string `json:"pass_results"`
}

type RunnerJSON struct {
	ContainerName    string   `json:"containerName"`
	ContainerTag     string   `json:"containerTag"`
	ContainerCommand []string `json:"containerCommand"`
}
