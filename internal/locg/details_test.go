package locg

import (
	"testing"
	"time"
)

func TestParseLoCGDate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Time
		errExpected bool
	}{
		{
			name:        "Valid date format",
			input:       "JAN 1, 2020",
			expected:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			errExpected: false,
		},
		{
			name:        "Valid date format with lowercase month",
			input:       "Feb 29, 2024",
			expected:    time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			errExpected: false,
		},
		{
			name:        "Invalid date format with full month name",
			input:       "January 12, 2021",
			expected:    time.Time{},
			errExpected: true,
		},
		{
			name:        "Invalid date format in different order",
			input:       "15 Mar 2022",
			expected:    time.Time{},
			errExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestExtractPrice(t *testing.T) {
	testCases := []struct {
		name          string
		rawPriceText  string
		expectedPrice int64
		expectedFound bool
	}{
		{
			name:          "Valid price at the end",
			rawPriceText:  "Comic · 32 pages · $3.99",
			expectedPrice: 399,
			expectedFound: true,
		},
		{
			name:          "Valid price with different currency",
			rawPriceText:  "Comic · 32 pages · $9.99",
			expectedPrice: 999,
			expectedFound: true,
		},
		{
			name:          "No price in string",
			rawPriceText:  "Comic · 32 pages · Rated T",
			expectedPrice: 0,
			expectedFound: false,
		},
		{
			name:          "Empty string",
			rawPriceText:  "",
			expectedPrice: 0,
			expectedFound: false,
		},
		{
			name:          "Price in the middle",
			rawPriceText:  "Comic · $4.99 · 32 pages",
			expectedPrice: 0,
			expectedFound: false,
		},
		{
			name:          "Only price",
			rawPriceText:  "$1.99",
			expectedPrice: 199,
			expectedFound: true,
		},
		{
			name:          "String with no separators",
			rawPriceText:  "Some random text",
			expectedPrice: 0,
			expectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			price, found := extractPrice(tc.rawPriceText)
			if price != tc.expectedPrice {
				t.Errorf("expected price to be '%v', but got '%v'", tc.expectedPrice, price)
			}
			if found != tc.expectedFound {
				t.Errorf("expected found to be '%t', but got '%t'", tc.expectedFound, found)
			}
		})
	}
}

func TestParsePrice(t *testing.T) {
	testCases := []struct {
		name          string
		priceText     string
		expectedPrice int64
		expectedFound bool
	}{
		{
			name:          "Valid price with dollar sign",
			priceText:     "$3.99",
			expectedPrice: 399,
			expectedFound: true,
		},
		{
			name:          "Invalid price without dollar sign",
			priceText:     "3.99",
			expectedPrice: 399,
			expectedFound: true,
		},
		{
			name:          "Invalid price with too many decimal places",
			priceText:     "$3.9123",
			expectedPrice: 391,
			expectedFound: true,
		},
		{
			name:          "Invalid price with non-numeric characters",
			priceText:     "$3.9a",
			expectedPrice: 0,
			expectedFound: false,
		},
		{
			name:          "Empty price string",
			priceText:     "",
			expectedPrice: 0,
			expectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			price, found := parsePrice(tc.priceText)
			if price != tc.expectedPrice {
				t.Errorf("expected price to be '%v', but got '%v'", tc.expectedPrice, price)
			}
			if found != tc.expectedFound {
				t.Errorf("expected found to be '%t', but got '%t'", tc.expectedFound, found)
			}
		})
	}
}
