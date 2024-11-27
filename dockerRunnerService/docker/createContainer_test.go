package docker

import (
	"strings"
	"testing"
)

func TestBasicHelloWorld(t *testing.T) {
	stdout, _, err := CreateContainer("hello-world", "latest", []string{}, []string{}, []string{}, false)
	if err != nil {
		t.Errorf("Error creating container: %v", err)
	}

	if stdout == "" {
		t.Errorf("Container output is empty")
	}

	if !strings.Contains(stdout, "Hello from Docker!") {
		t.Errorf("Container output is incorrect: %v", stdout)
	}
}
