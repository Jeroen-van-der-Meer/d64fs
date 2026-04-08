package cbm

import (
	"fmt"
	"io"
)

type registeredFormat struct {
	name  string
	sniff func(data []byte) bool
	open  func(data []byte) (Archive, error)
}

var formats []registeredFormat

// RegisterFormat registers a format for auto-detection by Open.
// Formats are tried in registration order; the first match wins.
func RegisterFormat(name string, sniff func([]byte) bool, open func([]byte) (Archive, error)) {
	formats = append(formats, registeredFormat{name, sniff, open})
}

// Open reads all data from r, auto-detects the format, and returns an Archive.
func Open(r io.Reader) (Archive, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}
	for _, f := range formats {
		if f.sniff(data) {
			return f.open(data)
		}
	}
	return nil, fmt.Errorf("unrecognized format (%d bytes)", len(data))
}
