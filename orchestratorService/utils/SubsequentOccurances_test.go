package utils

import "testing"

// Count occurrences of a string in a slice
func TestCountOccurrencesInSlice(t *testing.T) {
	slice := []string{"apple", "banana", "apple", "apple", "banana"}
	result := SubsequentOccurrences("apple", slice)
	expected := 2
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle an empty slice
func TestHandleEmptySliceOccurrences(t *testing.T) {
	slice := []string{}
	result := SubsequentOccurrences("apple", slice)
	expected := 0
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

func TestHandleSliceWithOneElement(t *testing.T) {
	slice := []string{"apple"}
	result := SubsequentOccurrences("apple", slice)
	expected := 1
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle a slice where the string appears at the end
func TestHandleSliceWhereStringAppearsAtEnd(t *testing.T) {
	slice := []string{"apple", "banana", "banana", "apple", "apple"}
	result := SubsequentOccurrences("apple", slice)
	expected := 2

	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle a slice where the string appears at the beginning
func TestSubsequentOccurrencesAtBeginning(t *testing.T) {
	slice := []string{"apple", "apple", "banana", "apple"}
	result := SubsequentOccurrences("apple", slice)
	expected := 2
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}

// Handle a slice with all elements being the same string
func TestAllElementsSameString(t *testing.T) {
	slice := []string{"apple", "apple", "apple", "apple"}
	result := SubsequentOccurrences("apple", slice)
	expected := 4
	if result != expected {
		t.Errorf("Expected %d, but got %d", expected, result)
	}
}
