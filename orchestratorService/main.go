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

	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		yamlFile, err := os.ReadFile("config.yaml")
		if err != nil {
			log.Fatalf("yamlFile.Get err   #%v ", err)
		}

		var jsonBody JSONbody
		err = json.NewDecoder(r.Body).Decode(&jsonBody)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// normalize the target
		jsonBody.Target, err = normalizeTarget(jsonBody.Target)
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

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// this function replaces the template arguments in the command arguments with set values like the target
func replaceTemplateArgs(args []string, target string) []string {
	for i, arg := range args {
		if arg == "{{req_domain}}" {
			args[i] = target
		}
	}
	return args
}

// this function kicks off a docker container with the given configuration and returns the output of the container
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
	resp, err := http.Post("http://"+os.Getenv("DOCKER_RUNNER_SERVICE"), "application/json", bytes.NewBuffer(jsonValue))
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

// this runs scans if the initials scans have vulnerabilities that require subsequent scans
// it runs the scans that are in the results map of the runner config
// TODO: send parsed results to decody
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
		if vulnerabilityPos != -1 { // if there is a match of the vulnerability in the results
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
	sr, err := runDockerService(rf)
	if err != nil {
		return "", err
	}

	pr := sendResultToParser(rf.ContainerName, sr)

	runSubsequentScans(pr, rf, t, cf)

	return sr, nil
}

func normalizeTarget(target string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("target is empty")
	}

	// strip the protocol from the target
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		target = strings.TrimPrefix(target, "http://")
		target = strings.TrimPrefix(target, "https://")
	}

	// strip the path from the target
	if strings.Contains(target, "/") {
		target = strings.Split(target, "/")[0]
	}

	return target, nil
}
