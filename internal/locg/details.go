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
	ReleaseDate time.Time
	CoverPrice  string
	Description string
	UPC         string
	URL         string
	StorageBox  string
}

// ScrapeComicBookDetails navigates to the given URL and extracts comic book details
func ScrapeComicBookDetails(parentCtx context.Context, item ComicItem) (*ComicBookDetails, error) {
	d := &ComicBookDetails{
		ID:  item.ID,
		URL: item.URL,
	}

	ctx, cancel := context.WithTimeout(parentCtx, 15*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(item.URL),

		fetchBasicData(d),
		fetchExtendedData(d),

		// STORAGE BOX: Need to click to "My Details" tab first
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
	fmt.Printf("ReleaseDate: %s\n", d.ReleaseDate.Format("02. Jan 2006"))
	fmt.Printf("Cover Price:   %s\n", d.CoverPrice)
	fmt.Printf("Description:   %s\n", d.Description)
	fmt.Printf("UPC:   %s\n", d.UPC)
	fmt.Printf("Box:       %s\n", d.StorageBox)
	fmt.Println("------------------------------------------------")

	return d, err
}

// fetchBasicData extracts the title, publisher, and UPC from the comic book details page
func fetchBasicData(d *ComicBookDetails) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// TITLE: h1 inside of div.page-details
		if err := chromedp.Text(`div.page-details h1`, &d.Title, chromedp.ByQuery).Do(ctx); err != nil {
			log.Printf("[Error] Failed to extract title: %v", err)
			return err
		}

		// PUBLISHER: Link inside of div.header-intro
		if err := chromedp.Text(`div.header-intro a`, &d.Publisher, chromedp.ByQuery).Do(ctx); err != nil {
			log.Printf("[Error] Failed to extract publisher: %v", err)
			return err
		}

		// UPC: XPath to find the UPC value
		// Search for div with class 'name' containing text 'UPC' and get the following sibling div with class 'value'
		// First check if the element exists to avoid errors when trying to extract text
		var nodes []*cdp.Node
		if err := chromedp.Nodes(`//div[contains(@class, 'name') and contains(text(), 'UPC')]/following-sibling::div[contains(@class, 'value')]`, &nodes, chromedp.AtLeast(0)).Do(ctx); err != nil {
			log.Printf("[Error] Failed to check for UPC element: %v)", err)
			return err
		}

		if len(nodes) > 0 {
			// Element exists, extract the text
			if err := chromedp.Text(`//div[contains(@class, 'name') and contains(text(), 'UPC')]/following-sibling::div[contains(@class, 'value')]`, &d.UPC, chromedp.BySearch).Do(ctx); err != nil {
				log.Printf("[Error] Failed to extract UPC: %v", err)
				return err
			}
		}
		return nil
	})
}

// fetchExtendedData extracts the release date from the comic book details page
func fetchExtendedData(d *ComicBookDetails) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var dateStr string
		dateSelector := `//div[contains(@class, 'header-intro')]//a[contains(@href, '/new-comics/')]`

		// RELEASE DATE: XPath to find the release date link inside the header intro
		var dateNodes []*cdp.Node
		if err := chromedp.Nodes(dateSelector, &dateNodes, chromedp.BySearch).Do(ctx); err == nil && len(dateNodes) > 0 {
			// Wichtig: chromedp.BySearch für XPath nutzen!
			if err := chromedp.Text(dateSelector, &dateStr, chromedp.BySearch).Do(ctx); err == nil {
				d.ReleaseDate, err = ParseLoCGDate(dateStr)
				if err != nil {
					log.Printf("[Error] Failed to parse release date: %v\n", err)
					return err
				}
			} else {
				log.Printf("[Error] Failed to extract release date: %v\n", err)
				return err
			}
		}

		// COVER PRICE: This is not very elegant, but it seems we have to be very precise to find the price element. It is inside a div.row with mt-1 and mb-4,
		// and then a div.col with classes copy-small and font-italic. We can try to find this element and extract the text, which should contain the price.
		priceSelector := `div.row.mt-1.mb-4 div.col.copy-small.font-italic`
		var rawPriceText string

		var priceNodes []*cdp.Node
		if err := chromedp.Nodes(priceSelector, &priceNodes, chromedp.ByQuery).Do(ctx); err == nil && len(priceNodes) > 0 {
			// This is something like "Comic · 32 pages · $3.99"
			chromedp.Text(priceSelector, &rawPriceText, chromedp.ByQuery).Do(ctx)
			log.Printf("[Info] Raw price text: %s\n", rawPriceText)
		} else {
			log.Printf("[Info] No price element found with selector: %s\n", priceSelector)
		}

		if price, ok := extractPrice(rawPriceText); ok {
			d.CoverPrice = price
		}

		// DESCRIPTION: div with class 'listing-description' inside of div.col12
		descriptionSelector := `div.col-12.listing-description`

		var descNodes []*cdp.Node
		if err := chromedp.Nodes(descriptionSelector, &descNodes, chromedp.ByQuery).Do(ctx); err == nil && len(descNodes) > 0 {
			chromedp.Text(descriptionSelector, &d.Description, chromedp.ByQuery).Do(ctx)
		}

		return nil
	})
}

// extractPrice tries to find a price string (e.g., "$3.99") from a raw text input.
// It assumes the price is the last component of a '·' separated string.
func extractPrice(rawPriceText string) (string, bool) {
	if rawPriceText == "" {
		return "", false
	}

	parts := strings.Split(rawPriceText, "·")
	// We are looking for the last part of the string.
	if len(parts) > 0 {
		lastPart := strings.TrimSpace(parts[len(parts)-1])
		if strings.HasPrefix(lastPart, "$") {
			return lastPart, true
		}
	}

	return "", false
}

// ParseLoCGDate converts a date string from LoCG format to time.Time
func ParseLoCGDate(dateStr string) (time.Time, error) {
	// LoCG dates are typically in the format "Month(Short) Day, Year" (e.g., "Jan 15, 2024")
	cleanStr := strings.TrimSpace(dateStr)
	layout := "Jan 2, 2006"
	parsedDate, err := time.Parse(layout, cleanStr)
	if err != nil {
		log.Printf("[Error] Failed to parse date '%s': %v\n", dateStr, err)
		return time.Time{}, err
	}
	return parsedDate, nil
}
