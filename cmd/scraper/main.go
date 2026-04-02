package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/chromedp/chromedp"
	"github.com/madhatter/comicbookcollector/internal/browser"
	"github.com/madhatter/comicbookcollector/internal/db"
	"github.com/madhatter/comicbookcollector/internal/locg"
	"github.com/madhatter/comicbookcollector/internal/ui"
)

// Configuration
const targetUsername = "nostalgix"
const checkURL = "https://leagueofcomicgeeks.com/settings"
const dbFile = "cbc.db"

func main() {
	// Initialize logging
	logFile, err := os.OpenFile("cbc.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Open database connection
	database, err := db.NewDatabase(dbFile)
	if err != nil {
		log.Fatalln("[FATAL] Failed to connect to database:", err)
	}
	defer database.Close()

	// Run database migrations to ensure the schema is up to date
	if err = database.Migrate(); err != nil {
		log.Fatalln("[FATAL] Failed to migrate database:", err)
	}

	// Initialize the progress UI
	p := tea.NewProgram(ui.NewProgressModel())
	done := make(chan struct{})

	go func() {
		p.Run()
		close(done)
	}()

	p.Send(ui.ComicScrapeMessage{
		StatusMessage: "Initializing... ",
	})

	// 1. Initialize Browser Session
	// headless = false so you can see the window
	sess, err := browser.NewSession(false)
	if err != nil {
		log.Fatalln("[FATAL] Failed to start browser session:", err)
	}
	defer sess.Close()

	// 2. Verify Login Status
	if err = sess.EnsureLoggedIn(targetUsername, checkURL); err != nil {
		p.Send(ui.ComicScrapeMessage{
			StatusMessage: "Connecting... Please log in manually via the browser window.",
		})

		log.Printf("[Warning] %v", err)
		log.Println("Please log in manually via the browser window.")
		log.Println("Waiting 60 seconds for manual login...")

		// Wait on the main browser context
		err = chromedp.Run(sess.Context, chromedp.Sleep(60*time.Second))

		if err != nil {
			log.Fatalln("[FATAL] Session died:", err)
		}
	}

	// 3. Get Collection Links
	log.Printf("[Info] Getting collection links for user '%s'...\n", targetUsername)

	p.Send(ui.ComicScrapeMessage{
		StatusMessage: "Loading Collection... ",
	})

	items, err := locg.GetCollectionLinks(sess.Context, targetUsername)
	if err != nil {
		log.Fatalln("[FATAL] Failed to get collection links:", err)
	}

	log.Printf("[Success] Found %d collection items.\n", len(items))

	// Load already-scraped IDs to skip them
	existingIDs, err := database.GetAllLoCGIDs()
	if err != nil {
		log.Fatalln("[FATAL] Failed to load existing IDs:", err)
	}
	existing := make(map[int]bool, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = true
	}

	var toScrape []locg.ComicItem
	for _, item := range items {
		if !existing[item.ID] {
			toScrape = append(toScrape, item)
		}
	}

	log.Printf("[Info] %d already in DB, %d to scrape.\n", len(existingIDs), len(toScrape))

	p.Send(ui.ComicScrapeMessage{
		Total:         len(toScrape),
		StatusMessage: "Starting scraper... ",
	})

	// 4. Scrape the details for each comic book in the collection
	for i, item := range toScrape {
		log.Printf("[%d/%d] Scrape: %s\n", i+1, len(toScrape), item.URL)

		select {
		case <-done:
			log.Println("Progress UI has finished. Stopping scraping.")
			return
		default:
		}

		details, err := locg.ScrapeComicBookDetails(sess.Context, item)
		if err != nil {
			log.Printf("-> ERROR: %v\n", err)
			if sess.Context.Err() != nil {
				log.Printf("[FATAL] Browser session lost: %v — stopping.\n", sess.Context.Err())
				break
			}
			continue
		}

		p.Send(ui.ComicScrapeMessage{
			Total:         len(toScrape),
			Current:       i + 1,
			StatusMessage: "Scraping... ",
			Detail: ui.ComicDetail{
				Title:       details.Title,
				IssueNumber: details.IssueNumber,
				Publisher:   details.Publisher,
				ReleaseDate: details.ReleaseDate,
				CoverPrice:  details.CoverPrice,
				Value:       details.Value,
				UPC:         details.UPC,
				Box:         details.StorageBox,
			},
		})

		// 5. Save the details to the database
		if err = database.SaveComic(details); err != nil {
			log.Printf("-> ERROR saving to database: %v\n", err)
			continue
		}

		log.Printf("-> Successfully saved to database.\n")

		// Random delay between 1-3 seconds to avoid bot detection
		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
	}
	p.Quit()
	<-done // Wait for the progress UI to finish
}
