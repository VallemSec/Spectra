package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"main/types"
	"main/utils"
)

var globalConfig types.ConfigFile
var log = logrus.New()

func init() {
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.TextFormatter{DisableColors: true}
	log.Out = os.Stderr

	log.Info("Initializing program")

	godotenv.Load()
	enforceEnvVars()
	log.Info("Loaded env vars")

	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Fatal(err)
	}
	log.Level = level
	log.Debugf("Set log level to %s", log.GetLevel())

	globalConfig, err = getAndUnmarshalConfigFile(os.Getenv("CONFIG_FILE_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Loaded config file")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", corsMiddleware(handleRequest))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var jsonBody types.JSONbody
	var prevScans []types.RunnerConfig

	err := json.NewDecoder(r.Body).Decode(&jsonBody)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if jsonBody.Email != "" {
		emailLeaks := checkEmailLeak(jsonBody.Email)
		json.NewEncoder(w).Encode(emailLeaks)
		return
	}

	jsonBody.Target, err = utils.NormalizeTarget(jsonBody.Target)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	decodyId := generateDecodyId(jsonBody.Target)
	requestLogger := log.WithField("DecodyId", decodyId)

	var wg sync.WaitGroup
	config := copyConfig()
	requestLogger.Trace("Created copy of the config")

	runRunnersConcurrently(config.DiscoveryRunners, config, jsonBody, decodyId, w, &wg, prevScans, requestLogger)
	runRunnersConcurrently(config.AlwaysRun, config, jsonBody, decodyId, w, &wg, prevScans, requestLogger)

	wg.Wait()

	resp, err := http.Get(os.Getenv("DECODY_SERVICE") + "/generate/" + decodyId)
	if err != nil {
		requestLogger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	io.Copy(w, resp.Body)
	requestLogger.Info("Finished running all scans")
}

func checkEmailLeak(email string) []types.EmailReturn {
	// call leakcheck.io with the email
	req, err := http.NewRequest("GET", "https://leakcheck.io/api/public", nil)
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("check", email)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)

	type LeakCheckResponse struct {
		Success bool     `json:"success"`
		Fields  []string `json:"fields"`
		Sources []struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"sources"`
	}

	var leakCheckResponse LeakCheckResponse
	err = json.NewDecoder(resp.Body).Decode(&leakCheckResponse)
	if err != nil {
		log.Fatal(err)
	}

	if leakCheckResponse.Success {
		var returnList []types.EmailReturn
		for _, field := range leakCheckResponse.Sources {
			returnList = append(returnList, types.EmailReturn{
				Service: field.Name,
				Date:    field.Date,
			})
		}

		return returnList
	}

	return []types.EmailReturn{}
}

func copyConfig() types.ConfigFile {
	// Errors are ignored since they should not show up since the object originates from a yaml file
	// which should already have stopped the program upon initializing if faulty
	var configCopy types.ConfigFile
	data, _ := yaml.Marshal(globalConfig)
	_ = yaml.Unmarshal(data, &configCopy)
	return configCopy
}

func enforceEnvVars() {
	envVariables := []string{"DOCKER_RUNNER_SERVICE", "CONFIG_FILE_PATH", "PARSERS_FOLDER", "PARSER_IMAGE", "PARSER_VERSION", "DECODY_SERVICE", "LOG_LEVEL"}
	defaultValueVariables := map[string]string{
		"DOCKER_RUNNER_SERVICE": "http://dockerrunner:8080",
		"PARSER_IMAGE":          "ghcr.io/vallemsec/spectra/parser",
		"PARSER_VERSION":        "latest",
		"DECODY_SERVICE":        "http://decody:5001",
		"LOG_LEVEL":             "warn",
	}

	for _, envVar := range envVariables {
		if os.Getenv(envVar) == "" && defaultValueVariables[envVar] == "" {
			log.Fatalf("%s environment variable is not set, exiting....", envVar)
		}
		if os.Getenv(envVar) == "" {
			if os.Setenv(envVar, defaultValueVariables[envVar]) != nil {
				log.Fatalf("could not set %s, exiting...", envVar)
			}
		}
	}
}

func runRunnersConcurrently(runners []string, config types.ConfigFile, jsonBody types.JSONbody, decodyId string, w http.ResponseWriter, wg *sync.WaitGroup, prevScans []types.RunnerConfig, logger *logrus.Entry) {
	for _, runnerName := range runners {
		wg.Add(1)
		go func(runnerName string) {
			defer wg.Done()
			runner := config.Runners[runnerName]

			fromConfig, err := runScan(runner, jsonBody.Target, decodyId, config, nil, prevScans, logger)
			if err != nil {
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			logger.Info("runFromConfig: ", fromConfig)
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

func getAndUnmarshalConfigFile(confPath string) (types.ConfigFile, error) {
	log.Debugf("Attempting to load %s", confPath)
	yamlFile, err := os.ReadFile(confPath)
	if err != nil {
		return types.ConfigFile{}, fmt.Errorf("could not read %s read config error: %v", confPath, err)
	}
	log.Debugf("Succesfully loaded %s", confPath)

	var config types.ConfigFile
	log.Debugf("Attempting to unmarshall the config file")
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return types.ConfigFile{}, fmt.Errorf("failed to unmarshall the config, this is typically due to a malformed config Unmarshalling error: %v", err)
	}
	log.Debugf("Succesfully unmarshall the config file")

	return config, nil
}

// runScan runs a scan from the configuration file
// if the scan has results that require subsequent scans, it runs the subsequent scans
// it returns the output of the scan
// it also protects against infinite recursion by keeping track of the scans that have been run and stopping if a scan has been run 3 times
// t is the target of the scan
func runScan(rf types.RunnerConfig, t, decodyId string, cf types.ConfigFile, res []string, prevScans []types.RunnerConfig, logger *logrus.Entry) (string, error) {
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
				result, err := runScan(rf, t, decodyId, cf, res, prevScans, logger)
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

	prevScans = append(prevScans, rf)

	if utils.SubsequentScanOccurrences(rf, prevScans) > 3 {
		return "", fmt.Errorf("scan has a loop %s, exiting. This happens when a scan is run 3 times without one in the middle", rf.ContainerName)
	}

	logger.Info("Running scan: ", rf.ContainerName)
	logger.Info("Args: ", rf.CmdArgs)
	sr, err := runDockerService(rf, []string{}, []string{}, logger)
	if err != nil {
		return "", err
	}

	joinedSr := strings.Join(sr, "\n")

	pr, err := sendResultToParser(rf, joinedSr, logger)
	if err != nil {
		return "", err
	}

	sendResultToDecody(pr, rf, decodyId, logger)

	runSubsequentScans(pr, rf, t, decodyId, cf, prevScans, logger)

	return joinedSr, nil
}

// this function kicks off a docker container with the given configuration and returns the output of the container
func runDockerService(runConf types.RunnerConfig, volumes, env []string, logger *logrus.Entry) ([]string, error) {
	logger.Info("Running docker service: ", runConf.Image, ":", runConf.ImageVersion, " with args: ", runConf.CmdArgs)
	configJSON := types.RunnerJSON{
		ContainerName:    runConf.Image,
		ContainerTag:     runConf.ImageVersion,
		ContainerCommand: runConf.CmdArgs,
		Volumes:          volumes,
		Env:              env,
		Tty:              runConf.Tty,
	}

	jsonValue, err := json.Marshal(configJSON)
	if err != nil {
		return []string{}, err
	}

	resp, err := http.Post(os.Getenv("DOCKER_RUNNER_SERVICE"), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return []string{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("docker runner service failed with status code %d", resp.StatusCode)
	}

	body := types.DockerRunnerResult{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		logger.Error("Error decoding response from docker runner service:", err)
		return []string{}, err
	}

	return body.Stdout, fmt.Errorf(strings.Join(body.Stderr, "\n"))
}

func sendResultToParser(runConf types.RunnerConfig, containerOutput string, logger *logrus.Entry) (types.ParserOutputJson, error) {
	// Clean the output of the container
	containerOutput = utils.CleanControlCharacters(containerOutput)

	serviceOut, err := runDockerService(types.RunnerConfig{
		Image:        os.Getenv("PARSER_IMAGE"),
		ImageVersion: os.Getenv("PARSER_VERSION"),
		CmdArgs:      []string{runConf.ContainerName, runConf.ParserPlugin, containerOutput},
	}, []string{os.Getenv("PARSERS_FOLDER") + ":/parsers"}, []string{"PARSER_FOLDER=/parsers"}, logger)
	if err != nil {
		// unmarshal the output of the parser error to get what's inside logger_name
		type errOutType struct {
			LoggerName string `json:"logger_name"`
			Message    string `json:"message"`
		}
		var errOut errOutType
		if err := json.Unmarshal([]byte(containerOutput), &errOut); err != nil {
			return types.ParserOutputJson{}, err
		}

		if errOut.LoggerName == "lua_panic_reporter" {
			logger.Error("Parser panic: ", errOut.Message)
			return types.ParserOutputJson{}, fmt.Errorf("parser error: %s", errOut.Message)
		} else if errOut.LoggerName == "lua_error_reporter" {
			logger.Error("Parser error: ", errOut.Message)
			return types.ParserOutputJson{}, fmt.Errorf("parser error: %s", errOut.Message)
		}

		return types.ParserOutputJson{}, err
	}

	serviceOutJoined := strings.Join(serviceOut, "\n")
	serviceOutJoined = utils.CleanParserOutput(serviceOutJoined)

	var pout types.ParserOutputJson
	if err := json.Unmarshal([]byte(serviceOutJoined), &pout); err != nil {
		logger.Error("Error unmarshalling parser output:", err)
		logger.Error("Parser output:", serviceOutJoined)
		return types.ParserOutputJson{}, err
	}

	return types.ParserOutputJson{
		ScannerName: runConf.ContainerName,
		Results:     pout.Results,
	}, nil
}

func sendResultToDecody(pout types.ParserOutputJson, rf types.RunnerConfig, decodyId string, logger *logrus.Entry) {
	if rf.Report == false || len(rf.DecodyRule) == 0 {
		return
	}

	decodyInput := types.DecodyInput{
		Name:    rf.ContainerName,
		Rules:   rf.DecodyRule,
		Results: pout.Results,
	}

	// marshal the results to send to decody
	jsonData, err := json.Marshal(decodyInput)
	if err != nil {
		logger.Errorln("Error marshalling decody input:", err)
		return
	}

	decodyUrl := os.Getenv("DECODY_SERVICE") + "/load/" + decodyId
	logger.Debugln("Decody load url:", decodyUrl)

	// send the results to decody
	res, err := http.Post(decodyUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorln("Error sending results to decody:", err)
	}

	// read the response from decody
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Errorln("Error reading response from decody:", err)
	}

	logger.Debug("Decody response: ", string(body))
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
func runSubsequentScans(pout types.ParserOutputJson, rc types.RunnerConfig, t, decodyId string, cf types.ConfigFile, prevScans []types.RunnerConfig, logger *logrus.Entry) {
	scansToRun := findScansToRun(pout, rc, cf)

	if scansToRun == nil {
		return
	}

	var wg sync.WaitGroup

	for _, result := range scansToRun {
		wg.Add(1)
		go func(result types.RunnerConfig) {
			defer wg.Done()
			runScan(result, t, decodyId, cf, pout.Results[0].PassRes, prevScans, logger)
		}(result)
	}

	// Wait for all scans to complete
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
