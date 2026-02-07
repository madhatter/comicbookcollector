package locg

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/madhatter/comicbookcollector/internal/browser"
)

type ComicItem struct {
	ID  int
	URL string
}

// GetCollectionLinks navigates to the user's collection page and extracts the links to individual comic book pages.
func GetCollectionLinks(ctx context.Context, username string) ([]ComicItem, error) {
	collectionURL := fmt.Sprintf("https://leagueofcomicgeeks.com/profile/%s/collection", username)
	log.Printf("[LoCG] Loading collection: %s\n", collectionURL)

	// We use a map to track seen URLs and avoid duplicates, as sometimes the site
	// might load duplicate items during infinite scrolling.
	seen := make(map[string]bool)
	var items []ComicItem

	// Every comic book in the collection is represented by an <a> tag with class "cover-link" inside a <li> in the "list-cover-grid".
	// Hopefully.
	linkSelector := `li.issue div.cover a`

	err := chromedp.Run(ctx,
		chromedp.Navigate(collectionURL),
		chromedp.Sleep(2*time.Second), // Wait for the page to load

		// Scroll to the bottom of the page to trigger loading of all items in the collection
		browser.ScrollToBottom(),

		// Extract the links to the comic book detail pages
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node

			if err := chromedp.Nodes(linkSelector, &nodes, chromedp.ByQueryAll).Do(ctx); err != nil {
				return err
			}

			log.Printf("[LoCG] Found nodes on collection side: %d\n", len(nodes))

			for _, n := range nodes {
				if n.AttributeValue("class") != "variant-toggle" {
					// Only process links that are not variant toggles, as those do not lead to comic detail pages.
					href := n.AttributeValue("href")
					if href != "" {
						id, err := ParseComicID(href)

						if err != nil {
							log.Printf("[Warning] %v\n", err)
							continue
						}

						// LoCG are relative (/comic/123...), we need to prepend the base URL
						fullURL := "https://leagueofcomicgeeks.com" + href
						if !seen[fullURL] {
							seen[fullURL] = true
							items = append(items, ComicItem{
								ID:  id,
								URL: fullURL,
							})
						} else {
							log.Printf("[Info] Skipping duplicate URL: %s\n", fullURL)
						}
					}
				}
				//log.Printf("   ...found comic link: %s\n", href)
			}
			return nil
		}),
	)

	return items, err
}

// ParseComicID extracts the LoCG comic ID from a given URL or path.
// Exptected: "/comic/12345/titel-slug" oder "https://league.../comic/12345/..."
// Returns: 12345
func ParseComicID(href string) (int, error) {
	parts := strings.Split(href, "/")

	// We look for the "comic" part in the URL and take the next part as the ID.
	for i, part := range parts {
		if part == "comic" && i+1 < len(parts) {
			id, err := strconv.Atoi(parts[i+1])
			if err != nil {
				return 0, fmt.Errorf("Failed to parse comic ID from URL %s: %v", href, err)
			}
			return id, nil
		}
	}
	return 0, fmt.Errorf("No ID found in URL: %s", href)
}
