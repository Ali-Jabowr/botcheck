package logparser

import (
	"fmt"
	"regexp"
)

type Entry struct {
	IP string
	UA string
}

func ParseLine(line string) (*Entry, error) {
	var re = regexp.MustCompile(`^(\S+) .*"([^"]*)"$`)

	matches := re.FindStringSubmatch(line)
	if len(matches) != 3 {
		return nil, fmt.Errorf("could not parse line: %s", line)
	}

	return &Entry{IP: matches[1], UA: matches[2]}, nil
}
