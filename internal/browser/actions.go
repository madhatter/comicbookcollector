package browser

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// ScrollToBottom scrolls to the bottom of the page, waiting for new content to load
// until it detects that no more content is being loaded (i.e., the scroll height stops increasing).
func ScrollToBottom() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var oldHeight, newHeight int

		for {
			// Get the current scroll height of the page
			if err := chromedp.Evaluate(`document.body.scrollHeight`, &newHeight).Do(ctx); err != nil {
				return err
			}

			log.Println("[LoCG] Scrolling...")

			// If the height hasn't changed, we have reached the bottom and can stop scrolling
			if newHeight == oldHeight {
				break
			}

			// Scroll to the bottom of the page
			if err := chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil).Do(ctx); err != nil {
				return err
			}

			// Wait for new content to load
			time.Sleep(2 * time.Second)

			oldHeight = newHeight
		}
		return nil
	})
}
