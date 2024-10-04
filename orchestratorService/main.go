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

	"main/types"
	"main/utils"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

var previousScans []string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file, exiting....")
	}

	// check if the DOCKER_RUNNER_SERVICE environment variable is set
	if os.Getenv("DOCKER_RUNNER_SERVICE") == "" {
		log.Fatal("DOCKER_RUNNER_SERVICE environment variable is not set, exiting....")
	} else if os.Getenv("CONFIG_FILE_PATH") == "" {
		log.Fatal("CONFIG_FILE_PATH environment variable is not set, exiting....")
	}

	configFileName := os.Getenv("CONFIG_FILE_PATH")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var jsonBody types.JSONbody
		var config types.ConfigFile
		yamlFile, err := os.ReadFile(configFileName)
		if err != nil {
			log.Fatalf("Could not read config.yaml read error: %v", err)
		}

		err = json.NewDecoder(r.Body).Decode(&jsonBody)
		jsonBody.Target, err = utils.NormalizeTarget(jsonBody.Target)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		err = yaml.Unmarshal(yamlFile, &config)
		if err != nil {
			log.Fatalf("Failed to unmarshall the config, this is typically due to a malformed config Unmarshalling error: %v", err)
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

// this function kicks off a docker container with the given configuration and returns the output of the container
func runDockerService(runConf types.RunnerConfig) (string, error) {
	configJSON := types.RunnerJSON{
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
func sendResultToParser(containerName, containerOutput string) types.ParserOutputJson {
	// send the result to the parser
	returnData := types.ParserOutputJson{
		ScannerName: containerName,
		Results: []types.Result{
			{
				Short: "HTTP",
				Long:  "{\\\"80\\\": {\\\"name\\\": \\\"http\\\"}",
			},
		},
	}

	return returnData
}

// runSubsequentScans runs scans if the initials scans have vulnerabilities that require subsequent scans
// it runs the scans that are in the results map of the runner config
// TODO: send parsed results to decody
func runSubsequentScans(pout types.ParserOutputJson, rc types.RunnerConfig, t string, cf types.ConfigFile) {
	// get all the keys of results
	var resultKeys []string
	for key, _ := range rc.Results {
		key = strings.ToUpper(key)
		resultKeys = append(resultKeys, key)
	}

	if resultKeys == nil {
		return
	}

	for _, vulnerability := range pout.Results {
		vulnerabilityPos := slices.Index(resultKeys, vulnerability.Short)
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

// runScanFromConfig runs a scan from the configuration file
// if the scan has results that require subsequent scans, it runs the subsequent scans
// it returns the output of the scan
// it also protects against infinite recursion by keeping track of the scans that have been run and stopping if a scan if it's been run 3 times
func runScanFromConfig(rf types.RunnerConfig, t string, cf types.ConfigFile) (string, error) {
	rf.CmdArgs = utils.ReplaceTemplateArgs(rf.CmdArgs, t)
	previousScans = append(previousScans, rf.ContainerName)

	// if the scan has been run 3 times after each other, stop the scan
	if utils.SubsequentOccurrences(rf.ContainerName, previousScans) > 3 {
		return "", fmt.Errorf("scan has a loop %s, exiting. This happens when a scan is run 3 times without one in the middle", rf.ContainerName)
	}

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
