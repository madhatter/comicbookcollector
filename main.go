package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const userDataDir = "./chrome-data"

func main() {
	// TODO: Use a folder in user home directory for user data
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	userDataDir := filepath.Join(wd, "chrome-data")

	fmt.Println("Using user data directory:", userDataDir)

	// Ensure the directory exists
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		if err := os.Mkdir(userDataDir, 0755); err != nil {
			log.Fatal("Could not create chrome-data directory:", err)
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
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Here starts the main logic

	// First check if we need to log in
	fmt.Println("[Info] Checking login status...")

	// Use a dummy URL that requires login or shows user-specific elements (e.g., your collection)
	checkURL := "https://leagueofcomicgeeks.com/profile/nostalgix/collection"

	var nodes []*cdp.Node

	err = chromedp.Run(ctx,
		chromedp.Navigate(checkURL),
		// Give the page a moment to render the DOM
		chromedp.Sleep(3*time.Second),
		// Check for the avatar image with your username in the alt tag.
		// Based on your HTML: <img ... alt="nostalgix">
		chromedp.Nodes(`img[alt='nostalgix']`, &nodes, chromedp.ByQuery),
	)

	if err != nil {
		log.Fatal(err)
	}

	// If no nodes found, we are not logged in
	if len(nodes) == 0 {
		fmt.Println("---------------------------------------------------------")
		fmt.Println("[Alert] NOT LOGGED IN (Avatar 'nostalgix' not found)!")
		fmt.Println("---------------------------------------------------------")
		fmt.Println("Browser window is open. Please log in manually.")
		fmt.Println("Waiting 60 seconds...")

		// Wait for manual login
		chromedp.Run(ctx, chromedp.Sleep(60*time.Second))
		fmt.Println("[Info] Resuming...")
	} else {
		fmt.Println("[Info] Login confirmed (Avatar found).")
	}

	// Now navigate to the target page to scrape data

	var (
		title     string
		publisher string
		upc       string
		box       string
	)

	// Example URL for now to scrape
	// TODO: Change to dynamic input if needed
	targetURL := "https://leagueofcomicgeeks.com/comic/5434313/nyx-4"

	fmt.Printf("[Info] Scraping: %s\n", targetURL)

	err = chromedp.Run(ctx,
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
