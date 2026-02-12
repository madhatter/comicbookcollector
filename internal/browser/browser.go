package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Session holds the browser context and cancellation functions.
type Session struct {
	Context     context.Context
	Cancel      context.CancelFunc
	allocCancel context.CancelFunc // Needed to properly close the allocator
}

// NewSession creates a new Chrome instance with a persistent user profile.
func NewSession(headless bool) (*Session, error) {
	// Determine the path for the user data directory
	// TODO: Use a folder in user home directory for user data
	wd, err := os.Getwd()
	if err != nil {
		return &Session{Context: nil, Cancel: nil, allocCancel: nil}, fmt.Errorf("unable to get working directory: %v", err)
	}

	// We construct an absolute path to avoid locking issues if run from different directories
	userDataDir := filepath.Join(wd, "chrome-data")

	// Ensure the directory exists
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		if err := os.Mkdir(userDataDir, 0o755); err != nil {
			return &Session{Context: nil, Cancel: nil, allocCancel: nil}, fmt.Errorf("unable to create user data directory: %v", err)
		}
	}

	// Configure Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// No headless mode at the start to allow manual login
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", false),
		// Force desktop resolution so menus don't collapse into mobile versions
		chromedp.WindowSize(1920, 1080),
		chromedp.UserDataDir(userDataDir),
	)

	// Create the allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create the browser context (without a timeout here, we handle timeouts per action)
	ctx, cancel := chromedp.NewContext(allocCtx)

	// Start the browser here to use the session immediately
	if err := chromedp.Run(ctx); err != nil {
		log.Println("[Browser] Unable to start: ", err)
		return &Session{Context: nil, Cancel: nil, allocCancel: nil}, err
	}

	return &Session{
		Context:     ctx,
		Cancel:      cancel,
		allocCancel: allocCancel,
	}, nil
}

// Close shuts down the browser and cleans up resources.
func (s *Session) Close() {
	s.Cancel()
	s.allocCancel()
}

// EnsureLoggedIn checks if the user is logged in and prompts for manual login if not.
func (s *Session) EnsureLoggedIn(username string, checkUrl string) error {
	//checkURL := fmt.Sprintf("https://leagueofcomicgeeks.com/profile/%s/collection", username)
	//settingsURL := "https://leagueofcomicgeeks.com/settings"
	log.Printf("[Browser] Verifying session for user '%s'...\n", username)
	log.Printf("[Browser] Testing login status at %s ...\n", checkUrl)

	var currentURL string

	// Use a short timeout for the check
	ctx, cancel := context.WithTimeout(s.Context, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(checkUrl),
		// Give the page a moment to render the DOM
		chromedp.Sleep(3*time.Second),
		chromedp.Location(&currentURL),
	)

	if err != nil {
		return err
	}

	// If redirected to login, user is not logged in
	if strings.Contains(currentURL, "/login") {
		return fmt.Errorf("user '%s' is not logged in", username)
	}

	log.Println("[Browser] Login confirmed.")
	return nil
}
