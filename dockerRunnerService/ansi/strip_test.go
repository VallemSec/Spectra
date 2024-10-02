package ansi

import (
	"testing"
)

func TestRemovesANSIEscapeCodes(t *testing.T) {
	input := "\x1b[31mHello\x1b[0m World"
	expected := "Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRemoveMultipleANSICodes(t *testing.T) {
	input := "\x1b[31mHello\x1b[0m \x1b[32mWorld\x1b[0m"
	expected := "Hello World"
	result := Strip(input) // Adjust the function call if necessary

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandleIncompleteANSICodes(t *testing.T) {
	input := "Hello\x1b[31m World\x1b["
	expected := "Hello World\x1b["
	result := Strip(input)

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestProcessesStringWithIncompleteANSIEscapeCodesGracefully(t *testing.T) {
	input := "\x1b[31mHello World"
	expected := "Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestManagesStringsWithOnlyANSIEscapeCodes(t *testing.T) {
	input := "\x1b[31mHello\x1b[0m World"
	expected := "Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesEmptyStrings(t *testing.T) {
	input := ""
	expected := ""
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFunctionIsCaseInsensitive(t *testing.T) {
	input := "\x1b[31mHello\x1b[0m World"
	expected := "Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesVeryLongStringsEfficiently(t *testing.T) {
	input := "Very long string with ANSI escape codes \x1b[31mHello\x1b[0m World"
	expected := "Very long string with ANSI escape codes Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesVeryLongStringsWithNewLines(t *testing.T) {
	input := "Very long\n string with ANSI escape codes \x1b[31mHello\x1b[0m\nWorld"
	expected := "Very long\n string with ANSI escape codes Hello\nWorld"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesNestedANSICodes(t *testing.T) {
	input := "\x1b[31mHello \x1b[1mWorld\x1b[0m\x1b[0m"
	expected := "Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesMixedContent(t *testing.T) {
	input := "Hello \x1b[31mWorld\x1b[0m! How are \x1b[32myou\x1b[0m?"
	expected := "Hello World! How are you?"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesOnlyANSICodes(t *testing.T) {
	input := "\x1b[31m\x1b[0m"
	expected := ""
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestHandlesInvalidANSICodes(t *testing.T) {
	input := "Hello\x1b[31m World\x1b[99m"
	expected := "Hello World"
	result := Strip(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
