package utils

import "testing"

// Count occurrences of a string present multiple times in the slice
func TestCountMultipleOccurrences(t *testing.T) {
	slice := []string{"apple", "banana", "apple", "orange", "apple"}
	s := "apple"
	expectedCount := 3

	actualCount := OccurrencesInSlice(s, slice)

	if actualCount != expectedCount {
		t.Errorf("expected %d, got %d", expectedCount, actualCount)
	}
}

// Handle an empty slice
func TestHandleEmptySlice(t *testing.T) {
	var slice []string
	s := "apple"
	expectedCount := 0

	actualCount := OccurrencesInSlice(s, slice)

	if actualCount != expectedCount {
		t.Errorf("expected %d, got %d", expectedCount, actualCount)
	}
}

// Return zero when the string is not present in the slice
func TestReturnZeroWhenStringNotPresent(t *testing.T) {
	slice := []string{"apple", "banana", "orange"}
	s := "grape"
	expectedCount := 0

	actualCount := OccurrencesInSlice(s, slice)

	if actualCount != expectedCount {
		t.Errorf("expected %d, got %d", expectedCount, actualCount)
	}
}

// Handle a large slice efficiently
func TestHandleLargeSliceEfficiently(t *testing.T) {
	// Prepare a large slice
	slice := make([]string, 1000)
	for i := range slice {
		slice[i] = "test"
	}

	s := "test"
	expectedCount := 1000

	// Test the function with a very large slice
	actualCount := OccurrencesInSlice(s, slice)

	if actualCount != expectedCount {
		t.Errorf("expected %d, got %d", expectedCount, actualCount)
	}
}

// Handle a slice with nil or empty strings
func TestHandleNilOrEmptyStrings(t *testing.T) {
	var slice []string
	s := "apple"
	expectedCount := 0

	actualCount := OccurrencesInSlice(s, slice)

	if actualCount != expectedCount {
		t.Errorf("expected %d, got %d", expectedCount, actualCount)
	}
}

// Handle a slice with all elements being the same as the target string
func TestAllElementsSameAsTargetString(t *testing.T) {
	slice := []string{"apple", "apple", "apple", "apple"}
	expectedCount := 4
	result := OccurrencesInSlice("apple", slice)
	if result != expectedCount {
		t.Errorf("Expected %d, but got %d", expectedCount, result)
	}
}
