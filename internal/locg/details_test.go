package locg

import (
	"testing"
	"time"
)

func TestParseLoCGDate(t *testing.T) {
	tests := []struct {
		input       string
		expected    time.Time
		errExpected bool
	}{
		{
			input:       "JAN 1, 2020",
			expected:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			errExpected: false,
		},
		{
			input:       "Feb 29, 2024",
			expected:    time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			errExpected: false,
		},
		{
			input:       "January 12, 2021",
			expected:    time.Time{},
			errExpected: true,
		},
		{
			input:       "15 Mar 2022",
			expected:    time.Time{},
			errExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseLoCGDate(tt.input)
			if (err != nil) != tt.errExpected {
				t.Errorf("ParseLoCGDate(%q) error = %v, expected error: %v", tt.input, err, tt.errExpected)
				return
			}
			if !tt.errExpected && !got.Equal(tt.expected) {
				t.Errorf("ParseLoCGDate(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
