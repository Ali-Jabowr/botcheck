package logparser

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	input := `192.0.2.1 - - [10/Oct/2026:13:55:36 +0000] "GET / HTTP/1.1" 200 1234 "-" "GPTBot/1.2"`
	entry, err := ParseLine(input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.IP != "192.0.2.1" {
		t.Errorf("IP = %q, want %q", entry.IP, "192.0.2.1")
	}

	if entry.UA != "GPTBot/1.2" {
		t.Errorf("UserAgnet = %q, want %q", entry.UA, "GPTBot/1.2")
	}

}
