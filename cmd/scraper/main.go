package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/madhatter/comicbookcollector/internal/browser"
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

	var (
		title     string
		publisher string
		upc       string
		box       string
	)

	log.Printf("[Info] Scraping: %s\n", targetURL)

	err := chromedp.Run(sess.Context,
		chromedp.Navigate(targetURL),

		// 1. TITEL: H1 inside div.page-details
		chromedp.Text(`div.page-details h1`, &title, chromedp.ByQuery),

		// 2. PUBLISHER: Link inside of div.header-intro
		chromedp.Text(`div.header-intro a`, &publisher, chromedp.ByQuery),

		// 3. UPC: XPath to find the UPC value
		// Search for div with class 'name' containing text 'UPC' and get the following sibling div with class 'value'
		chromedp.Text(`//div[contains(@class, 'name') and contains(text(), 'UPC')]/following-sibling::div[contains(@class, 'value')]`, &upc, chromedp.BySearch),

		// 4. STORAGE BOX: Need to click to "My Details" tab first
		chromedp.Click(`#nav-my-details-tab`, chromedp.ByQuery),

		// Wait for content to load
		chromedp.Sleep(2*time.Second),

		// Then get the storage box info
		chromedp.Value(`#storage_box`, &box, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatalf("[Error] Scraping failed: %v", err)
	}

	// Clean up data (trim newlines/spaces)
	title = strings.TrimSpace(title)
	publisher = strings.TrimSpace(publisher)
	upc = strings.TrimSpace(upc)
	box = strings.TrimSpace(box)

	fmt.Println("------------------------------------------------")
	fmt.Printf("Title:     %s\n", title)
	fmt.Printf("Publisher: %s\n", publisher)
	fmt.Printf("UPC:   %s\n", upc)
	fmt.Printf("Box:       %s\n", box)
	fmt.Println("------------------------------------------------")
}
