package main

import (
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/madhatter/comicbookcollector/internal/browser"
	"github.com/madhatter/comicbookcollector/internal/locg"
)

// Configuration
const targetUsername = "nostalgix"
const checkURL = "https://leagueofcomicgeeks.com/settings"

// Example URL for now to scrape
// TODO: Change to dynamic input if needed
const targetURL = "https://leagueofcomicgeeks.com/comic/5434313/nyx-4"

func main() {
	// 1. Initialize Browser Session
	// headless = false so you can see the window
	sess := browser.NewSession(false)
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

	// 3. Scrape Comic Book Details
	log.Printf("[Info] Scraping: %s\n", targetURL)
	details, err := locg.ScrapeComicBookDetails(sess.Context, targetURL)
	if err != nil {
		log.Fatalln("[FATAL] Scraping failed:", err)
	}

	log.Printf("[Success] Scraped details: %+v\n", details)
}
