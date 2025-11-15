package main

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFloat parses a TP-Link DDM DisplayString value to float64
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseFloat(s, 64)
}

// parsePort normalizes port string (extracts port from "1/0/N" format)
func parsePort(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty port")
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

