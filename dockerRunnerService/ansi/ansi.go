package ansi

import "regexp"

// Strip removes ANSI escape codes from a string
func Strip(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(input, "")
}
