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

## Prerequisites

- SQLite integration must be complete first (data needs to be persisted before it can be displayed)
