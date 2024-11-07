package utils

import "testing"

func CleanProperParserOutput(t *testing.T) {
	input := `["\u0001\u0000\u0000\u0000\u0000\u0000\u0000\ufffd{\"name\": \"nmap\", \"results\": [{\"short\": \"7070\", \"long\": \"realserver,84.247.14.132,7070,tcp\"}, {\"short\": \"80\", \"long\": \"http,84.247.14.132,80,tcp\"}, {\"short\": \"443\", \"long\": \"https,84.247.14.132,443,tcp\"}]}",""]`
	expected := `{"name": "nmap", "results": [{"long": "https,84.247.14.132,443,tcp", "short": "443"}, {"long": "http,84.247.14.132,80,tcp", "short": "80"}, {"long": "realserver,84.247.14.132,7070,tcp", "short": "7070"}]}`
	result := CleanParserOutput(input)

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func CleanParserOutputWithoutUnicodeChars(t *testing.T) {
	input := `["{\"name\": \"nmap\", \"results\": [{\"short\": \"7070\", \"long\": \"realserver,84.247.14.132,7070,tcp\"}, {\"short\": \"80\", \"long\": \"http,84.247.14.132,80,tcp\"}, {\"short\": \"443\", \"long\": \"https,84.247.14.132,443,tcp\"}]}",""]`
	expected := `{"name": "nmap", "results": [{"long": "https,84.247.14.132,443,tcp", "short": "443"}, {"long": "http,84.247.14.132,80,tcp", "short": "80"}, {"long": "realserver,84.247.14.132,7070,tcp", "short": "7070"}]}`

	result := CleanParserOutput(input)

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
