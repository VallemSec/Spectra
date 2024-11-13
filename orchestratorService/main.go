package main

import (
	"bytes"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"main/types"
	"main/utils"
)

var previousScans []string

func main() {
	initializeEnv()

	checkIfAllEnvVarsAreSet()

	config, err := getAndUnmarshalConfigFile(os.Getenv("CONFIG_FILE_PATH"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var jsonBody types.JSONbody
		decodyId := generateDecodyId(jsonBody.Target)

		err = json.NewDecoder(r.Body).Decode(&jsonBody)
		jsonBody.Target, err = utils.NormalizeTarget(jsonBody.Target)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		var wg sync.WaitGroup

		// Run DiscoveryRunners concurrently
		for _, runnerName := range config.DiscoveryRunners {
			wg.Add(1)
			go func(runnerName string) {
				defer wg.Done()
				runner := config.Runners[runnerName]

				fromConfig, err := runScanFromConfig(runner, jsonBody.Target, decodyId, config, nil)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}

				fmt.Println("runFromConfig: ", fromConfig)
			}(runnerName)
		}

		// Run AlwaysRun concurrently
		for _, runnerName := range config.AlwaysRun {
			wg.Add(1)
			go func(runnerName string) {
				defer wg.Done()
				runner := config.Runners[runnerName]

				fromConfig, err := runScanFromConfig(runner, jsonBody.Target, decodyId, config, nil)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}

				fmt.Println("runFromConfig: ", fromConfig)
			}(runnerName)
		}

		// Wait for all scans to complete
		wg.Wait()
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func initializeEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file, exiting....")
	}
}

func checkIfAllEnvVarsAreSet() {
	if os.Getenv("DOCKER_RUNNER_SERVICE") == "" {
		log.Fatal("DOCKER_RUNNER_SERVICE environment variable is not set, exiting....")
	} else if os.Getenv("CONFIG_FILE_PATH") == "" {
		log.Fatal("CONFIG_FILE_PATH environment variable is not set, exiting....")
	}
}

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

// runScanFromConfig runs a scan from the configuration file
// if the scan has results that require subsequent scans, it runs the subsequent scans
// it returns the output of the scan
// it also protects against infinite recursion by keeping track of the scans that have been run and stopping if a scan if it's been run 3 times
func runScanFromConfig(rf types.RunnerConfig, t, decodyId string, cf types.ConfigFile, res []string) (string, error) {
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
				result, err := runScanFromConfig(rf, t, decodyId, cf, res)
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
	sr, err := runDockerService(rf, []string{}, []string{})
	if err != nil {
		return "", err
	}

	pr := sendResultToParser(rf, sr)

	sendResultToDecody(sr, decodyId)

	runSubsequentScans(pr, rf, t, decodyId, cf)

	return sr, nil
}

// this function kicks off a docker container with the given configuration and returns the output of the container
func runDockerService(runConf types.RunnerConfig, volumes, env []string) (string, error) {
	fmt.Println("Running docker service: ", runConf.Image, ":", runConf.ImageVersion, " with args: ", runConf.CmdArgs)
	configJSON := types.RunnerJSON{
		ContainerName:    runConf.Image,
		ContainerTag:     runConf.ImageVersion,
		ContainerCommand: runConf.CmdArgs,
		Volumes:          volumes,
		Env:              env,
	}

	jsonValue, err := json.Marshal(configJSON)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(os.Getenv("DOCKER_RUNNER_SERVICE"), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func sendResultToParser(runConf types.RunnerConfig, containerOutput string) types.ParserOutputJson {
	// Clean the output of the container
	containerOutput = utils.CleanControlCharacters(containerOutput)

	serviceOut, err := runDockerService(types.RunnerConfig{
		Image:        "nekoluka/spectra-scanner",
		ImageVersion: "1.0.1",
		CmdArgs:      []string{runConf.ContainerName, runConf.ParserPlugin, containerOutput},
	}, []string{os.Getenv("PARSERS_FOLDER") + ":/parsers"}, []string{"PARSER_FOLDER=/parsers"})
	if err != nil {
		return types.ParserOutputJson{}
	}

	serviceOut = utils.CleanParserOutput(serviceOut)

	// parse the output of the parser
	var pout types.ParserOutputJson
	if err := json.Unmarshal([]byte(serviceOut), &pout); err != nil {
		log.Println("Error unmarshalling parser output:", err)
		log.Println("Parser output:", serviceOut)
		return types.ParserOutputJson{}
	}

	return types.ParserOutputJson{
		ScannerName: runConf.ContainerName,
		Results:     pout.Results,
	}
}

func sendResultToDecody(parsedOutput, decodyId string) {
	// send the results to decody
	_, err := http.Post(os.Getenv("DECODY_SERVICE"+"/"+decodyId), "application/json", bytes.NewBuffer([]byte(parsedOutput)))
	if err != nil {
		log.Println("Error sending results to decody:", err)
	}
}

// generateDecodyId makes a new unique identifier for the scan to send to decody based on the target and the current time
// id is in the format: base32(<target>)-<timestamp>
func generateDecodyId(target string) string {
	currentTime := time.Now().Unix()

	target = base32.StdEncoding.EncodeToString([]byte(target))

	return fmt.Sprintf("%s-%d", target, currentTime)
}

// runSubsequentScans runs scans if the initials scans have vulnerabilities that require subsequent scans
// it runs the scans that are in the results map of the runner config
func runSubsequentScans(pout types.ParserOutputJson, rc types.RunnerConfig, t, decodyId string, cf types.ConfigFile) {
	scansToRun := findScansToRun(pout, rc, cf)

	if scansToRun == nil {
		return
	}

	var wg sync.WaitGroup

	for _, result := range scansToRun {
		wg.Add(1)
		go func(result types.RunnerConfig) {
			defer wg.Done()
			runScanFromConfig(result, t, decodyId, cf, pout.Results[0].PassRes)
		}(result)
	}

	wg.Wait()
}

func findScansToRun(pout types.ParserOutputJson, rc types.RunnerConfig, cf types.ConfigFile) []types.RunnerConfig {
	var resultKeys []string
	for key := range rc.Results {
		key = strings.ToUpper(key)
		resultKeys = append(resultKeys, key)
	}

	if resultKeys == nil {
		return nil
	}

	scansToRun := make([]types.RunnerConfig, 0)
	seenScans := make(map[string]bool)

	for _, result := range pout.Results {
		for _, key := range resultKeys {
			if strings.Contains(key, "?") {
				if match, _ := regexp.MatchString(strings.ReplaceAll(key, "?", ".*"), result.Short); match {
					for _, scan := range rc.Results[key] {
						if !seenScans[scan] {
							scansToRun = append(scansToRun, cf.Runners[scan])
							seenScans[scan] = true
						}
					}
				}
			} else if key == result.Short {
				for _, scan := range rc.Results[key] {
					if !seenScans[scan] {
						scansToRun = append(scansToRun, cf.Runners[scan])
						seenScans[scan] = true
					}
				}
			}
		}
	}

	return scansToRun
}
