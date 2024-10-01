package utils

import (
	"testing"
)

func TestNormalizeTargetWorksWithValidTarget(t *testing.T) {
	target := "example.com"
	expected := "example.com"

	result, err := NormalizeTarget(target)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

// Handles target with only protocol and no additional characters
func TestNormalizeTargetRemovesHttpAndHttpsPrefix(t *testing.T) {
	target := "http://example.com"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	target = "https://example.com"
	result, err = NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetEmptyString(t *testing.T) {
	target := ""

	_, err := NormalizeTarget(target)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	expectedError := "target is empty"
	if err.Error() != expectedError {
		t.Errorf("expected error message %v, got %v", expectedError, err.Error())
	}
}

func TestNormalizeTargetRemovesPath(t *testing.T) {
	target := "example.com/path"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetRemovesMixedCaseProtocolPrefix(t *testing.T) {
	target := "hTtP://example.com"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	target = "HtTpS://example.com"
	result, err = NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesSubdomains(t *testing.T) {
	target := "sub.example.com"
	expected := "sub.example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesSubdomainsWithMixedCase(t *testing.T) {
	target := "SuB.ExAmPlE.cOm"
	expected := "sub.example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesSubdomainsWithMixedCaseProtocol(t *testing.T) {
	target := "hTtP://SuB.ExAmPlE.cOm"
	expected := "sub.example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}

	target = "HtTpS://SuB.ExAmPlE.cOm"
	result, err = NormalizeTarget(target)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesSubdomainsWithPath(t *testing.T) {
	target := "sub.example.com/path"
	expected := "sub.example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesSubdomainsWithPortNumbers(t *testing.T) {
	target := "sub.example.com:8080"
	expected := "sub.example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesSubdomainsWithQueryParameters(t *testing.T) {
	target := "sub.example.com?param=value"
	expected := "sub.example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesPortNumbers(t *testing.T) {
	target := "http://example.com:8080"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesPortNumbersWithPath(t *testing.T) {
	target := "example.com:8080/path"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesQueryParameters(t *testing.T) {
	target := "http://example.com?param=value"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesQueryParametersWithPath(t *testing.T) {
	target := "example.com?param=value/path"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestNormalizeTargetHandlesMultipleQueryParameters(t *testing.T) {
	target := "http://example.com?param1=value1&param2=value2"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

// Handles target with fragment identifiers
func TestNormalizeTargetHandlesFragmentIdentifiers(t *testing.T) {
	target := "http://example.com#section1"
	expected := "example.com"

	result, err := NormalizeTarget(target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
