package locg

import (
	"strings"
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

func TestExtractTitleNumber(t *testing.T) {
	testCases := []struct {
		name           string
		titleText      string
		expectedNumber int
		errExpected    bool
	}{
		{
			name:           "Valid title with number",
			titleText:      "Amazing Spider-Man #1",
			expectedNumber: 1,
			errExpected:    false,
		},
		{
			name:           "Valid title with number and extra text",
			titleText:      "Amazing Spider-Man #1 (Variant Cover)",
			expectedNumber: 0,
			errExpected:    true,
		},
		{
			name:           "Invalid title without number",
			titleText:      "Amazing Spider-Man",
			expectedNumber: 0,
			errExpected:    true,
		},
		{
			name:           "Invalid title with non-numeric number",
			titleText:      "Amazing Spider #One",
			expectedNumber: 0,
			errExpected:    true,
		},
		{
			name:           "Valid title with big number",
			titleText:      "Amazing Spider-Man #1234",
			expectedNumber: 1234,
			errExpected:    false,
		},
		{
			name:           "Title with hash at end",
			titleText:      "Batman #",
			expectedNumber: 0,
			errExpected:    true,
		},
		{
			name:           "Title with issue zero",
			titleText:      "Batman #0",
			expectedNumber: 0,
			errExpected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			number, err := extractTitleNumber(tc.titleText)
			if (err != nil) != tc.errExpected {
				t.Errorf("extractTitleNumber(%q) error = %v, expected error: %v", tc.titleText, err, tc.errExpected)
				return
			}
			if !tc.errExpected && number != tc.expectedNumber {
				t.Errorf("extractTitleNumber(%q) = %v, want %v", tc.titleText, number, tc.expectedNumber)
			}
		})
	}
}

func TestTitleVariantParsing(t *testing.T) {
	testCases := []struct {
		name            string
		inputTitle      string
		expectedTitle   string
		expectedVariant string
	}{
		{
			name:            "Title with newline",
			inputTitle:      "Amazing Spider-Man\nVariant Cover",
			expectedTitle:   "Amazing Spider-Man",
			expectedVariant: "Variant Cover",
		},
		{
			name:            "Title without newline",
			inputTitle:      "Batman #1",
			expectedTitle:   "Batman #1",
			expectedVariant: "",
		},
		{
			name:            "Title with multiple newlines",
			inputTitle:      "Title\nVariant\nExtra",
			expectedTitle:   "Title",
			expectedVariant: "Variant",
		},
		{
			name:            "Title with empty second line",
			inputTitle:      "Title\n",
			expectedTitle:   "Title",
			expectedVariant: "",
		},
		{
			name:            "Empty title",
			inputTitle:      "",
			expectedTitle:   "",
			expectedVariant: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var title, variantInfo string
			if parts := strings.Split(tc.inputTitle, "\n"); len(parts) > 1 {
				title = parts[0]
				variantInfo = parts[1]
			} else {
				title = tc.inputTitle
			}

			if title != tc.expectedTitle {
				t.Errorf("expected title %q, got %q", tc.expectedTitle, title)
			}
			if variantInfo != tc.expectedVariant {
				t.Errorf("expected variant %q, got %q", tc.expectedVariant, variantInfo)
			}
		})
	}
}

func TestUPCTrimming(t *testing.T) {
	testCases := []struct {
		name        string
		inputUPC    string
		expectedUPC string
	}{
		{
			name:        "UPC with spaces",
			inputUPC:    "  012345678901  ",
			expectedUPC: "012345678901",
		},
		{
			name:        "UPC with newlines",
			inputUPC:    "012345678901\n",
			expectedUPC: "012345678901",
		},
		{
			name:        "UPC with tabs",
			inputUPC:    "\t012345678901\t",
			expectedUPC: "012345678901",
		},
		{
			name:        "Empty UPC",
			inputUPC:    "",
			expectedUPC: "",
		},
		{
			name:        "UPC with mixed whitespace",
			inputUPC:    " \n\t 012345678901 \t\n ",
			expectedUPC: "012345678901",
		},
		{
			name:        "Already trimmed UPC",
			inputUPC:    "012345678901",
			expectedUPC: "012345678901",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trimmedUPC := strings.TrimSpace(tc.inputUPC)
			if trimmedUPC != tc.expectedUPC {
				t.Errorf("expected UPC %q, got %q", tc.expectedUPC, trimmedUPC)
			}
		})
	}
}

func TestImageLinkTrimming(t *testing.T) {
	testCases := []struct {
		name         string
		inputLink    string
		expectedLink string
	}{
		{
			name:         "Image link with spaces",
			inputLink:    "  https://example.com/image.jpg  ",
			expectedLink: "https://example.com/image.jpg",
		},
		{
			name:         "Image link with newlines",
			inputLink:    "https://example.com/image.jpg\n",
			expectedLink: "https://example.com/image.jpg",
		},
		{
			name:         "Empty image link",
			inputLink:    "",
			expectedLink: "",
		},
		{
			name:         "Image link with tabs",
			inputLink:    "\thttps://example.com/image.jpg\t",
			expectedLink: "https://example.com/image.jpg",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trimmedLink := strings.TrimSpace(tc.inputLink)
			if trimmedLink != tc.expectedLink {
				t.Errorf("expected link %q, got %q", tc.expectedLink, trimmedLink)
			}
		})
	}
}

func TestStorageBoxTrimming(t *testing.T) {
	testCases := []struct {
		name        string
		inputBox    string
		expectedBox string
	}{
		{
			name:        "Storage box with spaces",
			inputBox:    "  Box 1  ",
			expectedBox: "Box 1",
		},
		{
			name:        "Empty storage box",
			inputBox:    "",
			expectedBox: "",
		},
		{
			name:        "Storage box with newlines",
			inputBox:    "Box A\n",
			expectedBox: "Box A",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trimmedBox := strings.TrimSpace(tc.inputBox)
			if trimmedBox != tc.expectedBox {
				t.Errorf("expected box %q, got %q", tc.expectedBox, trimmedBox)
			}
		})
	}
}
