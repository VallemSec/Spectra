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
			log.Fatalf("yamlFile.Get err   #%v ", err)
		}

		jsonBody := JSONbody{}
		err = json.NewDecoder(r.Body).Decode(&jsonBody)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		var config ConfigFile
		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}

		for _, runnerName := range config.DiscoveryRunners {
			runner := config.Runners[runnerName]

			fromConfig, err := runScanFromConfig(runner, jsonBody.Target, config)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
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

func runDockerService(runConf RunnerConfig) (string, error) {
	type runnerJSON struct {
		ContainerName    string   `json:"containerName"`
		ContainerTag     string   `json:"containerTag"`
		ContainerCommand []string `json:"containerCommand"`
	}

	configJSON := runnerJSON{
		ContainerName:    runConf.Image,
		ContainerTag:     runConf.ImageVersion,
		ContainerCommand: runConf.CmdArgs,
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

	fmt.Println("Response body: ")
	fmt.Println(string(body))
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

func runSubsequentScans(pout ParserOutputJson, rc RunnerConfig, t string, cf ConfigFile) {
	// get all the keys of results
	var resultKeys []string
	for key, _ := range rc.Results {
		key = strings.ToUpper(key)
		resultKeys = append(resultKeys, key)
	}

	if resultKeys == nil {
		return
	}

	for _, vulnerability := range pout.Vulnerabilities {
		vulnerabilityPos := slices.Index(resultKeys, vulnerability.ErrShort)
		if vulnerabilityPos != -1 {
			// get the scans that need to be run
			scansToRun := rc.Results[resultKeys[vulnerabilityPos]]

			for _, scan := range scansToRun {
				runner := cf.Runners[scan]
				_, err := runScanFromConfig(runner, t, cf)
				if err != nil {
					return
				}
			}
		}
	}
}

func runScanFromConfig(rf RunnerConfig, t string, cf ConfigFile) (string, error) {
	rf.CmdArgs = replaceTemplateArgs(rf.CmdArgs, t)

	fmt.Println("Running scan: ", rf.ContainerName)
	fmt.Println("Args: ", rf.CmdArgs)
	serviceRes, err := runDockerService(rf)
	if err != nil {
		return "", err
	}

	parserRes := sendResultToParser(rf.ContainerName, serviceRes)

	runSubsequentScans(parserRes, rf, t, cf)

	return serviceRes, nil
}
