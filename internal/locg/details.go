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
		return nil
	})
}

// parseLoCGDate converts a date string from LoCG format to time.Time
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
