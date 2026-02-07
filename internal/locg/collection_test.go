package locg

import (
	"fmt"
	"testing"
)

func TestParseComicID(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		expected     int
		expectsError bool
	}{
		{
			name:         "Valid URL with ID",
			inputURL:     "https://leagueofcomicgeeks.com/comic/12345/some-comic-title",
			expected:     12345,
			expectsError: false,
		},
		{
			name:         "Valid URL with different path",
			inputURL:     "https://leagueofcomicgeeks.com/comic/67890/another-comic",
			expected:     67890,
			expectsError: false,
		},
		{
			name:         "Invalid URL format",
			inputURL:     "https://leagueofcomicgeeks.com/comic/some-comic-title",
			expected:     0,
			expectsError: true,
		},
		{
			name:         "Non-numeric ID",
			inputURL:     "https://leagueofcomicgeeks.com/comic/abcde/some-comic-title",
			expected:     0,
			expectsError: true,
		},
		{
			name:         "URL without comic segment",
			inputURL:     "https://leagueofcomicgeeks.com/series/12345/some-series-title",
			expected:     0,
			expectsError: true,
		},
		{
			name:         "Relative URL with ID",
			inputURL:     "/comic/54321/some-comic-title",
			expected:     54321,
			expectsError: false,
		},
		{
			name:         "Relative URL without ID",
			inputURL:     "/comic/some-comic-title",
			expected:     0,
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComicID(tt.inputURL)

			// Fehlerprüfung
			if (err != nil) != tt.expectsError {
				fmt.Printf("Error: %v, Expected Error: %v\n", err, tt.expectsError)
				t.Errorf("ParseComicID() error = %v, wantErr %v", err, tt.expectsError)
				return
			}

			// Ergebnisprüfung
			if got != tt.expected {
				t.Errorf("ParseComicID() = %v, want %v", got, tt.expected)
			}
		})
	}

}
