package main

import (
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/madhatter/comicbookcollector/internal/browser"
	"github.com/madhatter/comicbookcollector/internal/locg"
)

// Configuration
const targetUsername = "nostalgix"
const checkURL = "https://leagueofcomicgeeks.com/settings"

func main() {
	// 1. Initialize Browser Session
	// headless = false so you can see the window
	sess, err := browser.NewSession(false)
	if err != nil {
		log.Fatalln("[FATAL] Failed to start browser session:", err)
	}
	defer sess.Close()

	// 2. Verify Login Status
	if err := sess.EnsureLoggedIn(targetUsername, checkURL); err != nil {
		log.Printf("[Warning] %v", err)
		log.Println("Please log in manually via the browser window.")
		log.Println("Waiting 60 seconds for manual login...")

		// Wait on the main browser context
		err := chromedp.Run(sess.Context, chromedp.Sleep(60*time.Second))

		if err != nil {
			log.Fatalln("[FATAL] Session died:", err)
		}
	}

	// 3. Get Collection Links
	log.Printf("[Info] Getting collection links for user '%s'...\n", targetUsername)
	items, err := locg.GetCollectionLinks(sess.Context, targetUsername)
	if err != nil {
		log.Fatalln("[FATAL] Failed to get collection links:", err)
	}

	log.Printf("[Success] Found %d collection items.\n", len(items))

	// 4. Scrape the details for each comic book in the collection
	// For demonstration, we'll just scrape the first few items to avoid long runtimes during testing.
	//limit := min(len(items), 300)

	for i, item := range items {
		log.Printf("[%d/%d] Scrape: %s\n", i+1, len(items), item.URL)

		details, err := locg.ScrapeComicBookDetails(sess.Context, item)
		if err != nil {
			log.Printf("-> ERROR: %v\n", err)
			continue
		}

		fmt.Println("------------------------------------------------")
		fmt.Printf("Title:\t\t\t\t%s\n", details.Title)
		fmt.Printf("VariantInfo:\t\t\t%s\n", details.VariantInfo)
		fmt.Printf("Title Number:\t\t\t%d\n", details.TitleNumber)
		fmt.Printf("Publisher:\t\t\t%s\n", details.Publisher)
		fmt.Printf("ReleaseDate:\t\t\t%s\n", details.ReleaseDate.Format("02. Jan 2006"))
		fmt.Printf("Cover Price:\t\t\t$%.2f\n", float64(details.CoverPrice)/100.0)
		fmt.Printf("Description:\t\t\t%s\n", details.Description)
		fmt.Printf("Value:\t\t\t\t$%.2f\n", float64(details.Value)/100.0)
		fmt.Printf("ImageUrl:\t\t\t%s\n", details.ImageLink)
		fmt.Printf("UPC:\t\t\t\t%s\n", details.UPC)
		fmt.Printf("Box:\t\t\t\t%s\n", details.StorageBox)
		fmt.Println("------------------------------------------------")
	}
}
