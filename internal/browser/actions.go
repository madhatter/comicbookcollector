package browser

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// ScrollToBottom scrolls to the bottom of the page, waiting for new content to load
// until it detects that no more content is being loaded (i.e., the scroll height stops increasing).
func ScrollToBottom() chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		var oldHeight int

		for {
			// Get the current scroll height of the page
			newHeight, err := chromedp.Evaluate[int](`document.body.scrollHeight`)(ctx, t)
			if err != nil {
				return err
			}

			log.Println("[LoCG] Scrolling...")

			// If the height hasn't changed, we have reached the bottom and can stop scrolling
			if newHeight == oldHeight {
				break
			}

			// Scroll to the bottom of the page
			if _, err := chromedp.Evaluate[chromedp.Void](`window.scrollTo(0, document.body.scrollHeight)`)(ctx, t); err != nil {
				return err
			}

			// Wait for new content to load
			if _, err := chromedp.Sleep(2*time.Second)(ctx, t); err != nil {
				return err
			}

			oldHeight = newHeight
		}
		return nil
	})
}

// Into runs a and stores its result in dst, turning an Action[T] into an Action[Void].
func Into[T any](dst *T, a chromedp.Action[T]) chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
		v, err := a(ctx, t)
		if err != nil {
			return err
		}
		*dst = v
		return nil
	})
}
