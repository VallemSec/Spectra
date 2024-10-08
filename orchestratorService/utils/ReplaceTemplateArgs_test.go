package utils

import (
	"testing"
)

func TestReplaceTemplateArgsWithoutReplaceables(t *testing.T) {
	args := []string{"arg1", "arg2", "arg3"}
	target := "example.com"
	expected := [][]string{args}

	result := ReplaceTemplateArgs(args, target, nil)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	for i, v := range result {
		if len(v) != len(expected[i]) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		for j, arg := range v {
			if arg != expected[i][j] {
				t.Errorf("expected %v, got %v", expected, result)
			}
		}
	}
}

func TestReplaceTemplateArgsWithReqdomain(t *testing.T) {
	args := []string{"arg1", "{{req_domain}}", "arg3"}
	target := "example.com"
	expected := [][]string{{"arg1", "example.com", "arg3"}}

	result := ReplaceTemplateArgs(args, target, nil)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	for i, v := range result {
		if len(v) != len(expected[i]) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		for j, arg := range v {
			if arg != expected[i][j] {
				t.Errorf("expected %v, got %v", expected, result)
			}
		}
	}
}

func TestReplaceTemplateArgsWithPassString(t *testing.T) {
	args := []string{"arg1", "{{pass_results}}", "arg3"}
	target := "example.com"
	results := []string{"result1", "result2"}
	expected := [][]string{{"arg1", "result1 result2", "arg3"}}

	result := ReplaceTemplateArgs(args, target, results)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	for i, v := range result {
		if len(v) != len(expected[i]) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		for j, arg := range v {
			if arg != expected[i][j] {
				t.Errorf("expected %v, got %v", expected, result)
			}
		}
	}
}

func TestReplaceTemplateArgsWithPassArray(t *testing.T) {
	args := []string{"arg1", "{{[pass_results]}}", "arg3"}
	target := "example.com"
	results := []string{"result1", "result2"}
	expected := [][]string{{"arg1", "result1", "arg3"}, {"arg1", "result2", "arg3"}}

	result := ReplaceTemplateArgs(args, target, results)

	if len(result) != len(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	for i, v := range result {
		if len(v) != len(expected[i]) {
			t.Errorf("expected %v, got %v", expected, result)
		}

		for j, arg := range v {
			if arg != expected[i][j] {
				t.Errorf("expected %v, got %v", expected, result)
			}
		}
	}
}
