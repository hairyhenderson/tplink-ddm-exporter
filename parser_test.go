package tplinkddm

import (
	"testing"
)

func TestParseFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"temperature", "58.750000", 58.75, false},
		{"voltage", "3.290000", 3.29, false},
		{"bias current", "7.770000", 7.77, false},
		{"tx power", "-3.520000", -3.52, false},
		{"rx power", "-4.100000", -4.1, false},
		{"zero", "0.000000", 0, false},
		{"negative", "-10.500000", -10.5, false},
		{"invalid", "not a number", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseFloat(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseFloat() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("parseFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"single digit", "1", "1", false},
		{"double digit", "16", "16", false},
		{"with leading zero", "01", "1", false},
		{"tplink format", "1/0/2", "2", false},
		{"tplink format double digit", "1/0/16", "16", false},
		{"empty", "", "", true},
		{"non-numeric", "abc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parsePort(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parsePort() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if got != tt.want {
				t.Errorf("parsePort() = %v, want %v", got, tt.want)
			}
		})
	}
}
