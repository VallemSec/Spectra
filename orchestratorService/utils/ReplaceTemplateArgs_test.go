package utils

import (
	"testing"
)

func TestReplaceTemplateArgsWithoutReplaceables(t *testing.T) {
	args := []string{"arg1", "arg2", "arg3"}
	target := "example.com"
	expected := []string{"arg1", "arg2", "arg3"}

	result := ReplaceTemplateArgs(args, target)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("expected %v, got %v", expected, result)
		}
	}
}

func TestReplaceTemplateArgsWithReqdomain(t *testing.T) {
	args := []string{"arg1", "{{req_domain}}", "arg3"}
	target := "example.com"
	expected := []string{"arg1", "example.com", "arg3"}

	result := ReplaceTemplateArgs(args, target)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("expected %v, got %v", expected, result)
		}
	}
}
