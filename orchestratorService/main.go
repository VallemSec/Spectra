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
	"sync"

	"main/types"
	"main/utils"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

var previousScans []string

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file, exiting....")
	}

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

		if err := yaml.Unmarshal(yamlFile, &config); err != nil {
			log.Fatalf("Failed to unmarshall the config, this is typically due to a malformed config Unmarshalling error: %v", err)
		}

		for _, runnerName := range config.DiscoveryRunners {
			runner := config.Runners[runnerName]

			fromConfig, err := runScanFromConfig(runner, jsonBody.Target, config, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			log.Println(fromConfig)
		}
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// runScanFromConfig runs a scan from the configuration file
// if the scan has results that require subsequent scans, it runs the subsequent scans
// it returns the output of the scan
// it also protects against infinite recursion by keeping track of the scans that have been run and stopping if a scan if it's been run 3 times
func runScanFromConfig(rf types.RunnerConfig, t string, cf types.ConfigFile, res []string) (string, error) {
	replacedArgs := utils.ReplaceTemplateArgs(rf.CmdArgs, t, res)
	if len(replacedArgs) == 1 {
		rf.CmdArgs = replacedArgs[0]
	} else {
		var wg sync.WaitGroup
		var mu sync.Mutex
		var combinedResults []string
		var combinedErr error

		for _, arg := range replacedArgs {
			wg.Add(1)
			go func(arg []string) {
				defer wg.Done()
				rf.CmdArgs = arg
				result, err := runScanFromConfig(rf, t, cf, res)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					combinedErr = err
				} else {
					combinedResults = append(combinedResults, result)
				}
			}(arg)
		}
		wg.Wait()

		if combinedErr != nil {
			return "", combinedErr
		}
		return strings.Join(combinedResults, "\n"), nil
	}

	previousScans = append(previousScans, rf.ContainerName)

	if utils.SubsequentOccurrences(rf.ContainerName, previousScans) > 3 {
		return "", fmt.Errorf("scan has a loop %s, exiting. This happens when a scan is run 3 times without one in the middle", rf.ContainerName)
	}

	fmt.Println("Running scan: ", rf.ContainerName)
	fmt.Println("Args: ", rf.CmdArgs)
	sr, err := runDockerService(rf)
	if err != nil {
		return "", err
	}

	sr = `[
		{
			"short": "VULN-001",
			"long": "Description of vulnerability 001",
			"pass_results": "SinglePassResult"
		},
		{
			"short": "HTTP",
			"long": "{\\\"80\\\": {\\\"name\\\": \\\"http\\\"}",
			"pass_results": ["PassResult1", "PassResult2"]
		}
	]`

	pr := sendResultToParser(rf.ContainerName, sr)

	runSubsequentScans(pr, rf, t, cf)

	return sr, nil
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

	resp, err := http.Post("http://"+os.Getenv("DOCKER_RUNNER_SERVICE"), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Println("Response body: ")
	fmt.Println(string(body))

	return string(body), nil
}

func sendResultToParser(containerName, containerOutput string) types.ParserOutputJson {
	var results []types.Result
	if err := json.Unmarshal([]byte(containerOutput), &results); err != nil {
		log.Println("Error unmarshalling container output:", err)
		return types.ParserOutputJson{}
	}

	return types.ParserOutputJson{
		ScannerName: containerName,
		Results:     results,
	}
}

// runSubsequentScans runs scans if the initials scans have vulnerabilities that require subsequent scans
// it runs the scans that are in the results map of the runner config
// TODO: send parsed results to decody
func runSubsequentScans(pout types.ParserOutputJson, rc types.RunnerConfig, t string, cf types.ConfigFile) {
	var resultKeys []string
	for key := range rc.Results {
		key = strings.ToUpper(key)
		resultKeys = append(resultKeys, key)
	}

	if resultKeys == nil {
		return
	}

	for _, result := range pout.Results {
		vulnerabilityPos := slices.Index(resultKeys, result.Short)

		if vulnerabilityPos != -1 {
			scansToRun := rc.Results[resultKeys[vulnerabilityPos]]
			for _, scan := range scansToRun {
				runner := cf.Runners[scan]
				if _, err := runScanFromConfig(runner, t, cf, result.PassRes); err != nil {
					return
				}
			}
		}
	}
}
