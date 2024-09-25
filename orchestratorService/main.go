package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

type JSONbody struct {
	Target string `json:"target"`
}

type Config struct {
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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		yamlFile, err := os.ReadFile("config.yaml")
		if err != nil {
			log.Printf("yamlFile.Get err   #%v ", err)
			return
		}

		jsonBody := JSONbody{}
		err = json.NewDecoder(r.Body).Decode(&jsonBody)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}

		var config Config
		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}

		var discoveryRunners []RunnerConfig

		for _, runner := range config.DiscoveryRunners {
			discoveryRunners = append(discoveryRunners, config.Runners[runner])
		}

		// Use the config variable as needed
		for _, runner := range discoveryRunners {
			// replace the target in the cmdargs
			for i, arg := range runner.CmdArgs {
				if arg == "{{req_domain}}" {
					runner.CmdArgs[i] = jsonBody.Target
				}
			}

			dockerBody, err := runDockerService(runner)
			if err != nil {
				log.Println(err)
				// return the error as the response in JSON format
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"output": dockerBody})
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func runDockerService(config RunnerConfig) (string, error) {
	type runnerJSON struct {
		ContainerName    string   `json:"containerName"`
		ContainerTag     string   `json:"containerTag"`
		ContainerCommand []string `json:"containerCommand"`
	}

	configJSON := runnerJSON{
		ContainerName:    config.Image,
		ContainerTag:     config.ImageVersion,
		ContainerCommand: config.CmdArgs,
	}

	jsonValue, err := json.Marshal(configJSON)
	if err != nil {
		return "", err
	}

	// send the request to the dockerRunnerService
	resp, err := http.Post("http://localhost:8008", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	// read the response body
	body, err := io.ReadAll(resp.Body)
	log.Println(string(body))
	if err != nil {
		return "", err
	}

	// return the response body
	return string(body), nil
}
