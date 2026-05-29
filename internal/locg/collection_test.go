package locg

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func TestGetLinksSelector(t *testing.T) {
	// serve a fake HTML page that simulates the structure of the LoCG collection page
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
		<html><body>
			<li class="issue" data-comic="12345">
				<div class="cover"><a href="/comic/12345/batman">Link</a></div>
			</li>
			<li class="issue" data-comic="67890">
				<div class="cover"><a href="/comic/67890/superman">Link</a></div>
			</li>
		</body></html>
		`)
	}))
	defer ts.Close()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true), // Necessary for running in some CI environments and Docker containers
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-setuid-sandbox", true),
	)

	if chromePath := os.Getenv("CHROME_PATH"); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	// start a chromedp context
	ctx, cancelTimeout := context.WithTimeout(ctx, 120*time.Second)
	defer cancelTimeout()

	var nodes []*cdp.Node

	// This is the selector we use in GetCollectionLinks to find the comic links
	selector := `li.issue div.cover a`

	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll),
	)

	if err != nil {
		t.Fatalf("Chromedp error: %v", err)
	}

	// We expect to find 2 nodes based on our test HTML
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, but got %d", len(nodes))
	}
}

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
		{
			name:         "URL with query parameters",
			inputURL:     "https://leagueofcomicgeeks.com/comic/99999/comic-title?ref=collection",
			expected:     99999,
			expectsError: false,
		},
		{
			name:         "URL with trailing slash",
			inputURL:     "https://leagueofcomicgeeks.com/comic/11111/comic-title/",
			expected:     11111,
			expectsError: false,
		},
		{
			name:         "URL with query params and trailing slash",
			inputURL:     "https://leagueofcomicgeeks.com/comic/22222/comic-title/?ref=share",
			expected:     22222,
			expectsError: false,
		},
		{
			name:         "Empty URL",
			inputURL:     "",
			expected:     0,
			expectsError: true,
		},
		{
			name:         "URL with just root path",
			inputURL:     "https://leagueofcomicgeeks.com/",
			expected:     0,
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComicID(tt.inputURL)

			// Check if error presence matches expectation
			if (err != nil) != tt.expectsError {
				fmt.Printf("Error: %v, Expected Error: %v\n", err, tt.expectsError)
				t.Errorf("ParseComicID() error = %v, wantErr %v", err, tt.expectsError)
				return
			}

			// Check if the returned ID matches the expected value
			if got != tt.expected {
				t.Errorf("ParseComicID() = %v, want %v", got, tt.expected)
			}
		})
	}

}

func TestComicItemStruct(t *testing.T) {
	t.Run("ComicItem initialization", func(t *testing.T) {
		item := ComicItem{
			ID:  12345,
			URL: "https://leagueofcomicgeeks.com/comic/12345/test-comic",
		}

		if item.ID != 12345 {
			t.Errorf("Expected ID to be 12345, got %d", item.ID)
		}
		if item.URL != "https://leagueofcomicgeeks.com/comic/12345/test-comic" {
			t.Errorf("Expected URL to match, got %s", item.URL)
		}
	})

	t.Run("ComicItem zero value", func(t *testing.T) {
		var item ComicItem

		if item.ID != 0 {
			t.Errorf("Expected ID to be 0, got %d", item.ID)
		}
		if item.URL != "" {
			t.Errorf("Expected URL to be empty, got %s", item.URL)
		}
	})

	t.Run("ComicItem slice operations", func(t *testing.T) {
		items := []ComicItem{
			{ID: 1, URL: "https://example.com/comic/1/a"},
			{ID: 2, URL: "https://example.com/comic/2/b"},
			{ID: 3, URL: "https://example.com/comic/3/c"},
		}

		if len(items) != 3 {
			t.Errorf("Expected slice length 3, got %d", len(items))
		}

		if items[0].ID != 1 {
			t.Errorf("Expected first item ID to be 1, got %d", items[0].ID)
		}

		items = append(items, ComicItem{ID: 4, URL: "https://example.com/comic/4/d"})
		if len(items) != 4 {
			t.Errorf("Expected slice length 4 after append, got %d", len(items))
		}
	})
}
