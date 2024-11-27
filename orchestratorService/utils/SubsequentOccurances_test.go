package utils

import (
	"main/types"
	"testing"
)

var inputConfig = types.RunnerConfig{
	CmdArgs:       []string{"-t", "http://example.com"},
	Report:        true,
	ContainerName: "test",
	Image:         "test",
	ImageVersion:  "latest",
	ParserPlugin:  "test",
	DecodyRule:    []string{"test"},
	Results:       map[string][]string{"test": {"test"}},
}

var notInputConfig = types.RunnerConfig{
	CmdArgs:       []string{"-a", "http://vallem.com"},
	Report:        true,
	ContainerName: "not_test",
	Image:         "not_test",
	ImageVersion:  "latest",
	ParserPlugin:  "test",
	DecodyRule:    []string{"test"},
	Results:       map[string][]string{"test": {"test"}},
}

// Count occurrences of a string in a slice
func TestCountOccurrencesInSlice(t *testing.T) {
	slice := []types.RunnerConfig{
		inputConfig,
		notInputConfig,
		notInputConfig,
		inputConfig,
	}

	result := SubsequentScanOccurrences(inputConfig, slice)
	expected := 1
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle an empty slice
func TestHandleEmptySliceOccurrences(t *testing.T) {
	var slice []types.RunnerConfig
	result := SubsequentScanOccurrences(inputConfig, slice)
	expected := 0
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

func TestHandleSliceWithOneElement(t *testing.T) {
	slice := []types.RunnerConfig{inputConfig}
	result := SubsequentScanOccurrences(inputConfig, slice)
	expected := 1
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle a slice where the string appears at the end
func TestHandleSliceWhereStringAppearsAtEnd(t *testing.T) {
	slice := []types.RunnerConfig{notInputConfig, inputConfig, inputConfig, inputConfig}
	result := SubsequentScanOccurrences(inputConfig, slice)
	expected := 3

	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle a slice where the string appears at the beginning
func TestSubsequentOccurrencesAtBeginning(t *testing.T) {
	slice := []types.RunnerConfig{inputConfig, inputConfig, inputConfig, notInputConfig}
	result := SubsequentScanOccurrences(inputConfig, slice)
	expected := 0
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle a slice with all elements being the same string
func TestAllElementsSameString(t *testing.T) {
	slice := []types.RunnerConfig{inputConfig, inputConfig, inputConfig, inputConfig}
	result := SubsequentScanOccurrences(inputConfig, slice)
	expected := 4
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}
