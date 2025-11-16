package tplinkddm

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// parseFloat parses a TP-Link DDM DisplayString value to float64
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty string")
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse float: %w", err)
	}

	return f, nil
}

// parsePort normalizes port string (extracts port from "1/0/N" format)
func parsePort(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("empty port")
	}

	// Handle "1/0/N" format (stack/slot/port)
	parts := strings.Split(s, "/")
	if len(parts) == 3 {
		// Return just the port number
		return parts[2], nil
	}

	// Fallback: parse as int to remove leading zeros
	port, err := strconv.Atoi(s)
	if err != nil {
		return "", fmt.Errorf("invalid port: %w", err)
	}

	return strconv.Itoa(port), nil
}
