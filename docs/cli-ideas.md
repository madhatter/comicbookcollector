# CLI Ideas

A minimal CLI layer using [Cobra](https://github.com/spf13/cobra) — the de facto standard for Go CLIs (used by kubectl, gh, hugo, and many others). Cobra provides Unix-standard flag syntax (`--flag`, `-f`) unlike the built-in `flag` package.

## Subcommand structure

```
comicbookcollector            → launches the TUI (default, no subcommand)
comicbookcollector scrape     → runs the scraper
```

## File layout

```
cmd/
  main.go       ← entry point, registers subcommands via cobra
  scrape.go     ← scrape subcommand
  tui.go        ← tui subcommand (later)
```

## Progress bar during scraping

The `scrape` subcommand should show a progress bar in the terminal while scraping. Bubbles (v2) ships a ready-made `progress` component that fits well here — even without a full TUI. Since `len(items)` is known before the scraping loop starts, the bar can be driven deterministically (increment by 1 per comic).

This would be a good first hands-on use of the Bubble Tea / Bubbles ecosystem before building the full TUI.

## Next steps

- Finish SQLite persistence layer first (see `docs/database-ideas.md`)
- Then introduce Cobra and restructure `cmd/`
- Add progress bar as part of the `scrape` subcommand
