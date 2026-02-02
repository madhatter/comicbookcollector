package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const userDataDir = "./chrome-data"

var firstRun bool = false

func main() {
	// Check if user data directory exists, if not create it
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		firstRun = true

		err := os.Mkdir("./chrome-data", 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Set Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// No headless mode at the start to allow manual login
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		// Set user data dir to persist session
		chromedp.UserDataDir(userDataDir),
	)

	// Create allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create Chrome instance
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set a timeout to avoid hanging
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Here starts the main logic

	var (
		title     string
		publisher string
		upc       string
		box       string
	)

	// Example URL for now to scrape
	// TODO: Change to dynamic input if needed
	targetURL := "https://leagueofcomicgeeks.com/comic/5434313/nyx-4"

	fmt.Println("Starting Chrome...")

	err := chromedp.Run(ctx,
		// Navigieren
		chromedp.Navigate(targetURL),

		// If it's the first run, wait longer for manual login
		func() chromedp.Action {
			if firstRun {
				fmt.Println("First time start: Please log in within 120 seconds...")
				return chromedp.Sleep(120 * time.Second)
			} else {
				fmt.Println("Waiting for page to load...")
				return chromedp.Sleep(3 * time.Second)
			}
		}(),

		// 1. TITEL: H1 inside div.page-details
		chromedp.Text(`div.page-details h1`, &title, chromedp.ByQuery),

		// 2. PUBLISHER: Link inside of div.header-intro
		chromedp.Text(`div.header-intro a`, &publisher, chromedp.ByQuery),

		// 3. UPC: XPath to find the UPC value
		// Search for div with class 'name' containing text 'UPC' and get the following sibling div with class 'value'
		chromedp.Text(`//div[contains(@class, 'name') and contains(text(), 'UPC')]/following-sibling::div[contains(@class, 'value')]`, &upc, chromedp.BySearch),

		// 4. BOX: Jetzt müssen wir den Tab wechseln!
		// Wir klicken auf den Tab "My Details"
		chromedp.Click(`#nav-my-details-tab`, chromedp.ByQuery),

		// Warten, bis der Tab-Inhalt geladen ist (wichtig!)
		chromedp.Sleep(2*time.Second),

		chromedp.Value(`#storage_box`, &box, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal(err)
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
