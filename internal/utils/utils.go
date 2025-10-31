package utils

import "fmt"

// ParseInt parses a string parameter as an integer.
// Returns an error if the string cannot be parsed as an integer.
func ParseInt(s string) (int, error) {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return 0, err
	}
	return val, nil
}

