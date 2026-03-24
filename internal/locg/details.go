package locg

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type ComicBookDetails struct {
	ID          int
	VariantInfo string
	Title       string
	Series      string
	IssueNumber int
	Publisher   string
	ReleaseDate time.Time
	CoverPrice  int64 // Stored in cents
	Value       int64 // Stored in cents
	Description string
	UPC         string
	URL         string
	StorageBox  string
	ImageLink   string
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
		chromedp.Sleep(1*time.Second),

		// Then get the storage box info
		chromedp.Value(`#storage_box`, &d.StorageBox, chromedp.ByQuery),

		// MARKET VALUE: Extract the actual value of the comic book from the "Market" tab.
		chromedp.Click(`#nav-market-tab`, chromedp.ByQuery),

		// Wait for content to load
		chromedp.Sleep(1*time.Second),

		// The value of the comic book is found in the "Market" tab. We prioritize the "My Comic" value,
		// as it is most relevant to the user's collection. If it's not available, we fall back to the "Raw" value.
		chromedp.ActionFunc(func(ctx context.Context) error {
			var valueStr string
			myComicSelector := `//div[text()='My Comic' and contains(@class, 'copy-large')]/parent::div/following-sibling::div/div[contains(@class, 'copy-large')]`
			rawSelector := `//div[text()='Raw' and contains(@class, 'copy-large')]/parent::div/following-sibling::div/div[contains(@class, 'copy-large')]`

			// First, try to get the "My Comic" value.
			err := chromedp.Text(myComicSelector, &valueStr, chromedp.BySearch).Do(ctx)

			// If "My Comic" value is not found, is empty, or is just a placeholder, try the "Raw" value.
			if err != nil || strings.TrimSpace(valueStr) == "" || strings.TrimSpace(valueStr) == "-" {
				if err != nil {
					log.Printf("[Info] Could not find 'My Comic' value: %v. Falling back to 'Raw'.", err)
				} else {
					log.Printf("[Info] 'My Comic' value is empty or a placeholder ('%s'). Falling back to 'Raw'.", valueStr)
				}

				if err2 := chromedp.Text(rawSelector, &valueStr, chromedp.BySearch).Do(ctx); err2 != nil {
					log.Printf("[Info] Could not find 'Raw' value either: %v", err2)
					return nil // Don't block, just log it.
				}
			}

			if priceCents, ok := parsePrice(valueStr); ok {
				d.Value = priceCents
			}

			return nil
		}),

		// TODO: Click on "Series" and extract the series name?
	)

	if err != nil {
		log.Println("[Error] Scraping failed: ", err)
		return nil, err
	}

	// Clean up data (trim newlines/spaces)
	d.Title = strings.TrimSpace(d.Title)
	d.VariantInfo = strings.TrimSpace(d.VariantInfo)
	d.Publisher = strings.TrimSpace(d.Publisher)
	d.UPC = strings.TrimSpace(d.UPC)
	d.Description = strings.TrimSpace(d.Description)
	d.ImageLink = strings.TrimSpace(d.ImageLink)
	d.StorageBox = strings.TrimSpace(d.StorageBox)

	return d, err
}

// fetchBasicData extracts the title, publisher, and UPC from the comic book details page
func fetchBasicData(d *ComicBookDetails) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// TITLE: h1 inside of div.page-details
		var title string
		if err := chromedp.Text(`div.page-details h1`, &title, chromedp.ByQuery).Do(ctx); err != nil {
			log.Printf("[Error] Failed to extract title: %v", err)
			return err
		}

		if parts := strings.Split(title, "\n"); len(parts) > 1 {
			d.Title = parts[0]
			d.VariantInfo = parts[1]
		} else {
			d.Title = title
		}

		// Try to extract the title number from the title string. This is not always possible, so we log an error if it fails.
		var err2 error
		if d.IssueNumber, err2 = extractTitleNumber(d.Title); err2 != nil {
			log.Printf("[Error] Failed to extract title number: %s\n", err2)
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

		if priceCents, ok := extractPrice(rawPriceText); ok {
			d.CoverPrice = priceCents
		}

		// DESCRIPTION: div with class 'listing-description' inside of div.col12
		descriptionSelector := `div.col-12.listing-description`

		var descNodes []*cdp.Node
		if err := chromedp.Nodes(descriptionSelector, &descNodes, chromedp.ByQuery).Do(ctx); err == nil && len(descNodes) > 0 {
			chromedp.Text(descriptionSelector, &d.Description, chromedp.ByQuery).Do(ctx)
		}

		// IMAGE LINK: div with class 'cover-art' and then a link with class 'cover-gallery'
		imgSelector := `//div[contains(@class, 'cover-art')]/a[contains(@class, 'cover-gallery')]`

		var imgNodes []*cdp.Node
		if err := chromedp.Nodes(imgSelector, &imgNodes, chromedp.BySearch).Do(ctx); err == nil && len(imgNodes) > 0 {
			var exists bool
			if err := chromedp.AttributeValue(imgSelector, "href", &d.ImageLink, &exists, chromedp.BySearch).Do(ctx); err == nil && exists {
				log.Printf("[Info] Image link: %s\n", d.ImageLink)
			} else {
				log.Printf("[Error] Failed to extract image link: %v\n", err)
			}
		}
		return nil
	})
}

// extractValueNumber extracts the numeric value from a string like "Issue #12".
func extractTitleNumber(titleStr string) (int, error) {
	cleanStr := strings.TrimSpace(titleStr)
	parts := strings.Split(cleanStr, "#")
	if len(parts) < 2 {
		return 0, errors.New("title does not contain an issue number")
	}
	numberStr := strings.TrimSpace(parts[len(parts)-1])
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		log.Printf("[Error] Could not parse issue number from string '%s': %v", numberStr, err)
		return 0, err
	}
	return number, nil
}

// parsePrice converts a price string (e.g., "$3.99") into an integer in cents (e.g., 399).
func parsePrice(priceStr string) (int64, bool) {
	if priceStr == "" {
		return 0, false
	}
	cleanStr := strings.TrimSpace(priceStr)
	cleanStr = strings.TrimPrefix(cleanStr, "$")
	floatVal, err := strconv.ParseFloat(cleanStr, 64)
	if err != nil {
		log.Printf("[Error] Could not parse price string '%s' to float: %v", cleanStr, err)
		return 0, false
	}
	// Multiply by 100 and add 0.5 for proper rounding before converting to int64
	totalCents := int64(floatVal*100 + 0.5)
	return totalCents, true
}

// extractPrice tries to find a price string (e.g., "$3.99") from a raw text input
// that might contain other information (e.g., "Comic · 32 pages · $3.99").
// It assumes the price is the last component of a '·' separated string.
func extractPrice(rawPriceText string) (int64, bool) {
	if rawPriceText == "" {
		return 0, false
	}

	parts := strings.Split(rawPriceText, "·")
	// We are looking for the last part of the string.
	if len(parts) > 0 {
		priceStr := strings.TrimSpace(parts[len(parts)-1])
		if strings.HasPrefix(priceStr, "$") {
			return parsePrice(priceStr)
		}
	}
	return 0, false
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
