package locg

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type ComicBookDetails struct {
	ID          int
	Title       string
	Series      string
	IssueNumber int
	Publisher   string
	UPC         string
	URL         string
	StorageBox  string
}

// ScrapeComicBookDetails navigates to the given URL and extracts comic book details
func ScrapeComicBookDetails(parentCtx context.Context, item ComicItem) (ComicBookDetails, error) {
	var d ComicBookDetails
	d.URL = item.URL
	d.ID = item.ID

	var nodes []*cdp.Node

	ctx, cancel := context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(item.URL),

		// 1. TITEL: H1 inside div.page-details
		chromedp.Text(`div.page-details h1`, &d.Title, chromedp.ByQuery),

		// 2. PUBLISHER: Link inside of div.header-intro
		chromedp.Text(`div.header-intro a`, &d.Publisher, chromedp.ByQuery),

		// 3. UPC: XPath to find the UPC value
		// Search for div with class 'name' containing text 'UPC' and get the following sibling div with class 'value'
		// First check if the element exists to avoid errors when trying to extract text
		chromedp.Nodes(`//div[contains(@class, 'name') and contains(text(), 'UPC')]/following-sibling::div[contains(@class, 'value')]`, &nodes, chromedp.AtLeast(0)),

		chromedp.ActionFunc(func(ctx context.Context) error {
			if len(nodes) > 0 {
				// Element ist da, jetzt können wir den Text holen
				return chromedp.Text(`//div[contains(@class, 'name') and contains(text(), 'UPC')]/following-sibling::div[contains(@class, 'value')]`, &d.UPC, chromedp.BySearch).Do(ctx)
			}
			log.Println("Element does not exist, skipping...")
			return nil
		}),

		// 4. STORAGE BOX: Need to click to "My Details" tab first
		chromedp.Click(`#nav-my-details-tab`, chromedp.ByQuery),

		// Wait for content to load
		chromedp.Sleep(2*time.Second),

		// Then get the storage box info
		chromedp.Value(`#storage_box`, &d.StorageBox, chromedp.ByQuery),

		// TODO: Add issue number extraction if needed (not currently implemented
		// TODO: Click on "Series" and extract the series name?
	)

	if err != nil {
		log.Fatalf("[Error] Scraping failed: %v", err)
	}

	// Clean up data (trim newlines/spaces)
	d.Title = strings.TrimSpace(d.Title)
	d.Publisher = strings.TrimSpace(d.Publisher)
	d.UPC = strings.TrimSpace(d.UPC)
	d.StorageBox = strings.TrimSpace(d.StorageBox)

	fmt.Println("------------------------------------------------")
	fmt.Printf("Title:     %s\n", d.Title)
	fmt.Printf("Publisher: %s\n", d.Publisher)
	fmt.Printf("UPC:   %s\n", d.UPC)
	fmt.Printf("Box:       %s\n", d.StorageBox)
	fmt.Println("------------------------------------------------")

	return d, err
}
