package main

import (
	"bytes"
	"crypto/sha256"
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
	godotenv.Load()

	checkIfAllEnvVarsAreSet()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var jsonBody types.JSONbody

		// TODO: Read the config once and make copies when replacing args in runScan DO NOT MODIFY THE CONFIG ONCE IT IS READ
		config, err := getAndUnmarshalConfigFile(os.Getenv("CONFIG_FILE_PATH"))

		err = json.NewDecoder(r.Body).Decode(&jsonBody)
		jsonBody.Target, err = utils.NormalizeTarget(jsonBody.Target)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		decodyId := generateDecodyId(jsonBody.Target)

		var wg sync.WaitGroup

		// Run DiscoveryRunners concurrently
		runRunnersConcurrently(config.DiscoveryRunners, config, jsonBody, decodyId, w, &wg)

		// Run AlwaysRun concurrently
		runRunnersConcurrently(config.AlwaysRun, config, jsonBody, decodyId, w, &wg)

		// Wait for all scans to complete
		wg.Wait()

		// get the results from decody
		resp, err := http.Get(os.Getenv("DECODY_SERVICE") + "/generate/" + decodyId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// return the results from decody to the client
		io.Copy(w, resp.Body)

		fmt.Println("Finished running all scans")

	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func runRunnersConcurrently(runners []string, config types.ConfigFile, jsonBody types.JSONbody, decodyId string, w http.ResponseWriter, wg *sync.WaitGroup) {
	for _, runnerName := range runners {
		wg.Add(1)
		go func(runnerName string) {
			defer wg.Done()
			runner := config.Runners[runnerName]

			fromConfig, err := runScan(runner, jsonBody.Target, decodyId, config, nil)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			fmt.Println("runFromConfig: ", fromConfig)
		}(runnerName)
	}
}

func checkIfAllEnvVarsAreSet() {
	envVariables := []string{"DOCKER_RUNNER_SERVICE", "CONFIG_FILE_PATH", "PARSERS_FOLDER", "PARSER_IMAGE", "PARSER_VERSION", "DECODY_SERVICE"}

	for _, envVar := range envVariables {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s environment variable is not set, exiting....", envVar)
		}
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

// runScan runs a scan from the configuration file
// if the scan has results that require subsequent scans, it runs the subsequent scans
// it returns the output of the scan
// it also protects against infinite recursion by keeping track of the scans that have been run and stopping if a scan has been run 3 times
func runScan(rf types.RunnerConfig, t, decodyId string, cf types.ConfigFile, res []string) (string, error) {
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
				result, err := runScan(rf, t, decodyId, cf, res)
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

	sendResultToDecody(pr, rf, decodyId)

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
		Image:        os.Getenv("PARSER_IMAGE"),
		ImageVersion: os.Getenv("PARSER_VERSION"),
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

func sendResultToDecody(parsedOutput types.ParserOutputJson, rf types.RunnerConfig, decodyId string) {
	if rf.Report == false || len(rf.DecodyRule) == 0 {
		return
	}

	decodyInput := types.DecodyInput{
		Name:    rf.ContainerName,
		Rules:   rf.DecodyRule,
		Results: parsedOutput.Results,
	}

	// marshal the results to send to decody
	jsonData, err := json.Marshal(decodyInput)
	if err != nil {
		log.Println("Error marshalling decody input:", err)
		return
	}

	fmt.Println(os.Getenv("DECODY_SERVICE") + "/load/" + decodyId)

	// send the results to decody
	res, err := http.Post(os.Getenv("DECODY_SERVICE")+"/load/"+decodyId, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending results to decody:", err)
	}

	// read the response from decody
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response from decody:", err)
	}

	fmt.Println("Decody response: ", string(body))
}

// generateDecodyId makes a new unique identifier for the scan to send to decody based on the target and the current time
// id is in the format: base32(<target>)-<timestamp>
func generateDecodyId(target string) string {
	currentTime := time.Now().Unix()

	h := sha256.New()
	h.Write([]byte(target))
	target = fmt.Sprintf("%x", h.Sum(nil))

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
			runScan(result, t, decodyId, cf, pout.Results[0].PassRes)
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
