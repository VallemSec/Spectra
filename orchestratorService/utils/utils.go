package utils

import (
	"fmt"
	"strings"
)

// ReplaceTemplateArgs replaces the template arguments in the command arguments with the target.
func ReplaceTemplateArgs(args []string, target string) []string {
	for i, arg := range args {
		if arg == "{{req_domain}}" {
			args[i] = target
		}
	}
	return args
}

// NormalizeTarget normalizes the target by stripping the protocol and path from the target.
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
