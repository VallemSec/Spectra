package utils

import (
	"fmt"
	"main/types"
	"regexp"
	"strings"
	"sync"
)

// ReplaceTemplateArgs replaces the template arguments in the command arguments with the target.
// {{req_domain}} is replaced with the target.
// {{pass_results}} is replaced with the first result in the results array and makes a long space seperated string out of all the results.
// {{[pass_results]}} will return an array of args to run multiple scans with the results.
func ReplaceTemplateArgs(args []string, t string, res []string) [][]string {
	willPassAmount := len(res)
	willRunMultipleScans := false

	// replace the target in the command arguments
	for i, arg := range args {
		if arg == "{{req_domain}}" {
			args[i] = t
		}
		if arg == "{{pass_results}}" {
			args[i] = strings.Join(res, " ")
		}
		if arg == "{{[pass_results]}}" {
			args[i] = ""
			willRunMultipleScans = true
		}
	}

	// if there are multiple results, create multiple command arguments
	if willPassAmount > 1 && willRunMultipleScans {
		var wg sync.WaitGroup
		newArgs := make([][]string, willPassAmount)
		for i := 0; i < willPassAmount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				newArgs[i] = make([]string, len(args))
				copy(newArgs[i], args)
				for j, arg := range newArgs[i] {
					if arg == "" {
						newArgs[i][j] = res[i]
					}
				}
			}(i)
		}
		wg.Wait()
		return newArgs
	}

	return [][]string{args}
}

// NormalizeTarget normalizes the target by stripping the protocol and path from the target.
// The target is returned in lowercase and must not be empty.
func NormalizeTarget(target string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("target is empty")
	}

	// make sure the target all lowercase
	target = strings.ToLower(target)

	// strip the protocol from the target
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		target = strings.TrimPrefix(target, "http://")
		target = strings.TrimPrefix(target, "https://")
	}

	// strip the path from the target
	if strings.Contains(target, "/") {
		target = strings.Split(target, "/")[0]
	}

	// strip the port from the target
	if strings.Contains(target, ":") {
		target = strings.Split(target, ":")[0]
	}

	// strip the query from the target
	if strings.Contains(target, "?") {
		target = strings.Split(target, "?")[0]
	}

	// strip the fragment from the target
	if strings.Contains(target, "#") {
		target = strings.Split(target, "#")[0]
	}

	return target, nil
}

// SubsequentScanOccurrences counts the number of times a runner config appears in a slice of runner configs.
// please only use this function after replacing the template arguments in the runner config
func SubsequentScanOccurrences(rc types.RunnerConfig, scansRan []types.RunnerConfig) int {
	subsequentScans := 0

	for _, scanRan := range scansRan {
		if runnerConfigEqualish(rc, scanRan) {
			subsequentScans++
		} else {
			subsequentScans = 0
		}
	}
	return subsequentScans
}

// runnerConfigEqualish compares two runner configs and returns true if they are nearly the same
// for this we compare the name, container image, and the cmd args given to the container
func runnerConfigEqualish(a, b types.RunnerConfig) bool {
	strikes := 0

	// Compare the name, container image
	if len(a.ContainerName) == len(b.ContainerName) {
		strikes++
	}
	if len(a.Image) == len(b.Image) {
		strikes++
	}

	// compare the configs
	sameConfigArs := 0

	if len(a.CmdArgs) == len(b.CmdArgs) {
		for i := range a.CmdArgs {
			if a.CmdArgs[i] == b.CmdArgs[i] {
				sameConfigArs++
			}
		}
		if sameConfigArs == len(a.CmdArgs) {
			strikes++
		}
	}

	// if we have 3 strikes, we can say the runner configs are equalish
	if strikes >= 3 {
		return true
	}

	return false
}

func CleanParserOutput(input string) string {
	input = CleanControlCharacters(input)
	input = removeEscapeChars(input)

	input = strings.ReplaceAll(input, `,""`, ``)
	input = input[2 : len(input)-3]

	return input
}

func CleanControlCharacters(input string) string {
	// TODO: Actully make this work instead of hardcoding the control characters
	// Define a regular expression to match control characters, including Unicode control characters
	re := regexp.MustCompile(`(?im)\\u[a-z]*[0-9]*[a-z]?\*?#?`)
	// Replace all control characters with an empty string
	return re.ReplaceAllString(input, "")
}

func removeEscapeChars(input string) string {
	// Define a regular expression to match escape characters
	re := regexp.MustCompile(`\\`)
	// Replace all escape characters with an empty string
	return re.ReplaceAllString(input, "")
}
