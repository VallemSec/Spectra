package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"

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

type ParserOutputJson struct {
	ScannerName     string          `json:"scanner_name"`
	Vulnerabilities []vulnerability `json:"vulnerabilities"`
}

type vulnerability struct {
	ErrShort string `json:"err_short"`
	ErrLong  string `json:"err_long"`
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

		for _, runnerName := range config.DiscoveryRunners {
			runner := config.Runners[runnerName]

			fromConfig, err := runScanFromConfig(runner, jsonBody.Target, config)
			if err != nil {
				return
			}

			log.Println(fromConfig)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func replaceTemplateArgs(args []string, target string) []string {
	for i, arg := range args {
		if arg == "{{req_domain}}" {
			args[i] = target
		}
	}
	return args
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

// TODO: implement the function without sample data
func sendResultToParser(containerName, containerOutput string) ParserOutputJson {
	// send the result to the parser

	returnData := ParserOutputJson{
		ScannerName: containerName,
		Vulnerabilities: []vulnerability{
			{
				ErrShort: "HTTP",
				ErrLong:  "{\\\"80\\\": {\\\"name\\\": \\\"http\\\"}",
			},
		},
	}

	return returnData
}

func runSubsequentScans(parserOutput ParserOutputJson, config RunnerConfig, target string, configFile Config) {
	// get all the keys of results
	var resultKeys []string
	for key, _ := range config.Results {
		key = strings.ToUpper(key)
		resultKeys = append(resultKeys, key)
	}

	for _, vulnerability := range parserOutput.Vulnerabilities {
		fmt.Println(vulnerability)
		fmt.Println(resultKeys)
		vulnerabilityPos := slices.Index(resultKeys, vulnerability.ErrShort)
		fmt.Println(vulnerabilityPos)
		if vulnerabilityPos != -1 {
			// get the scans that need to be run
			scansToRun := config.Results[resultKeys[vulnerabilityPos]]

			for _, scan := range scansToRun {
				runner := configFile.Runners[scan]
				fromConfig, err := runScanFromConfig(runner, target, configFile)
				if err != nil {
					return
				}

				log.Println(fromConfig)
			}
		}
	}
}

func runScanFromConfig(config RunnerConfig, target string, configFile Config) (string, error) {
	config.CmdArgs = replaceTemplateArgs(config.CmdArgs, target)

	serviceRes, err := runDockerService(config)
	if err != nil {
		return "", err
	}

	parserRes := sendResultToParser(config.ContainerName, serviceRes)

	runSubsequentScans(parserRes, config, target, configFile)

	return serviceRes, nil
}
