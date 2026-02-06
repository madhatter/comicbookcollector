package locg

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// GetCollectionLinks navigates to the user's collection page and extracts the links to individual comic book pages.
func GetCollectionLinks(ctx context.Context, username string) ([]string, error) {
	collectionURL := fmt.Sprintf("https://leagueofcomicgeeks.com/profile/%s/collection", username)
	log.Printf("[LoCG] Loading collection: %s\n", collectionURL)

	var comicLinks []string

	// Every comic book in the collection is represented by an <a> tag with class "cover-link" inside a <li> in the "list-cover-grid".
	// Hopefully.
	linkSelector := `li.issue div.cover a`

	err := chromedp.Run(ctx,
		chromedp.Navigate(collectionURL),
		chromedp.Sleep(2*time.Second), // Wait for the page to load

		// INFINITE SCROLL SIMULATION
		// TODO: For now only scroll 5 times, in case of very large collections we might want to scroll more or detect when we've reached the end.
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 5; i++ { // scroll 5x
				fmt.Printf("   ...scrolling (%d/5)\n", i+1)
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
					// LoCG are relative (/comic/123...), we need to prepend the base URL
					// TODO: In the future we might want to extract the comic ID from the URL and store it separately, but for now let's just keep the full URL.
					fullURL := "https://leagueofcomicgeeks.com" + href
					comicLinks = append(comicLinks, fullURL)
				}
			}
			return nil
		}),
	)

	return comicLinks, err
}
