package cli

import (
	"bufio"
	"strings"
)

// readChoice reads a line, trims, lowercases, and returns the normalized choice
// Never returns an error for normal console usage; callers can ignore the second value
func readChoice(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(line)), nil
}
