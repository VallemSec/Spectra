package docker

import (
	"strings"
	"testing"
)

func TestBasicHelloWorld(t *testing.T) {
	container, err := CreateContainer("hello-world", "latest", []string{}, []string{}, []string{})
	if err != nil {
		t.Errorf("Error creating container: %v", err)
	}

	if container == "" {
		t.Errorf("Container output is empty")
	}

	if !strings.Contains(container, "Hello from Docker!") {
		t.Errorf("Container output is incorrect: %v", container)
	}
}
