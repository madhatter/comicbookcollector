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

		// INFINITE SCROLL SIMULATION
		// TODO: For now only scroll 5 times, in case of very large collections we might want to scroll more or detect when we've reached the end.
		chromedp.ActionFunc(func(ctx context.Context) error {
			//for i := 0; i < 5; i++ { // scroll 5x
			for i := range [5]int{} { // scroll 5x
				log.Printf("   ...scrolling (%d/5)\n", i+1)
				_ = chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil).Do(ctx)
				time.Sleep(3 * time.Second) // wait for new content to load, site seems to load quite slow
			}
			return nil
		}),

		// Extract the links to the comic book detail pages
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node

			if err := chromedp.Nodes(linkSelector, &nodes, chromedp.ByQueryAll).Do(ctx); err != nil {
				return err
			}

			log.Printf("[LoCG] Found items: %d\n", len(nodes))

			for _, n := range nodes {
				href := n.AttributeValue("href")
				if href != "" {
					// Extract the comic ID from the URL
					parts := strings.Split(href, "/")
					var id int

					// Make sure that the URL has the expected format before trying to extract the ID
					if len(parts) >= 3 {
						var err error
						id, err = strconv.Atoi(parts[2])
						if err != nil {
							log.Printf("[Warning] Failed to parse comic ID from URL %s: %v\n", href, err)
							continue
						}
					} else {
						// Fallback, if the URL format is unexpected, we can log a warning and skip this item
						log.Printf("[Warning] URL is not in expected format: %s\n", href)
						continue
					}

					// LoCG are relative (/comic/123...), we need to prepend the base URL
					// TODO: In the future we might want to extract the comic ID from the URL and store it separately, but for now let's just keep the full URL.
					fullURL := "https://leagueofcomicgeeks.com" + href
					if !seen[fullURL] {
						seen[fullURL] = true
						items = append(items, ComicItem{
							ID:  id,
							URL: fullURL,
						})
					}
				}
				//log.Printf("   ...found comic link: %s\n", href)
			}
			return nil
		}),
	)

	return items, err
}
