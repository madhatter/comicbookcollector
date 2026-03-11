# TUI / CLI Ideas

A terminal UI for browsing and managing the local comic book collection, once the SQLite persistence layer is in place.

## Library candidates

- [Bubble Tea v2](https://charm.land/blog/v2/) — the leading TUI framework in the Go ecosystem, follows the Elm architecture (Model/Update/View)
- **Lip Gloss v2** — styling and layout for terminal output, pairs naturally with Bubble Tea
- **Bubbles v2** — ready-made UI components (lists, tables, text inputs, spinners, …)

All three are from [Charm](https://charm.sh) and were released as v2 in early 2026 with significant performance improvements and a more declarative API.

## Possible features

- Browse the collection as a scrollable, filterable list
- View comic details (cover image, description, market value, storage box)
- Mark comics as read / update storage box
- Basic stats (total value, issues per publisher, …)

## Scraper improvements

### Progress bar during scraping
Bubbles (v2) ships a ready-made `progress` component. Since the total number of comics is known before scraping starts (`len(items)`), the bar can be driven deterministically — increment by 1 after each comic is processed. Could be added to the scraper standalone, without needing a full TUI first.

### Skip already-known comics
Before scraping detail pages, fetch all existing `locg_id` values from the DB into a map, then skip any comic whose ID is already present. Requires a `GetAllIDs() ([]int, error)` function on `Database`. Good first use case for Goroutines later (see below).

### Goroutines for parallel scraping
Scraping 1000+ comics sequentially is slow (each page has a 15s timeout). Goroutines could parallelize the detail fetching — but needs careful rate limiting to avoid being blocked by LoCG, and coordinating multiple chromedp contexts is non-trivial. Good learning project for Go concurrency once the basics are solid.

## Prerequisites

- SQLite integration must be complete first (data needs to be persisted before it can be displayed)
- A CLI layer (Cobra) will be introduced before the TUI — see `docs/cli-ideas.md`
