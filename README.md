# ComicBookCollector

![Build Status](https://github.com/madhatter/comicbookcollector/actions/workflows/cicd.yml/badge.svg)
![Go Version](https://img.shields.io/github/go-mod/go-version/madhatter/comicbookcollector)

**ComicBookCollector** is a specialized CLI tool written in Go to automate the extraction of comic book collection data from [League of Comic Geeks](https://leagueofcomicgeeks.com).

It uses a headless browser approach to handle dynamic content, infinite scrolling, and authenticated sessions, ensuring you get a complete backup of your collection metadata locally.

## 🚀 Current Features

* **Browser Automation:** Uses `chromedp` to control a real Chrome instance (headless or visible).
* **Smart Authentication:** Detects login state. If 2FA or Captchas are required, it pauses and waits for user interaction before proceeding.
* **Infinite Scroll Handling:** Automatically scrolls through the collection list until all items are loaded in the DOM.
* **Robust Scraping:**
    * Extracts **Base Data**: Title, Publisher, UPC.
    * Extracts **Extended Metadata**: Release Date, Cover Price, Plot Description.
    * Extracts **Private Data**: Custom "Storage Box" information (requires UI interaction/clicking tabs).
* **CI/CD Pipeline:** Automated builds and tests for Linux and macOS (AMD64/ARM64) via GitHub Actions.

## 🛠️ Tech Stack

* **Language:** Go (Golang) 1.23+
* **Core Library:** [chromedp](https://github.com/chromedp/chromedp) (Chrome DevTools Protocol)
* **Architecture:** Modular design separating browser logic (`internal/browser`) from business logic (`internal/locg`).

## 📦 Installation & Usage

### Prerequisites
* Go installed
* Google Chrome installed

### Running the Scraper

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/madhatter/comicbookcollector.git](https://github.com/madhatter/comicbookcollector.git)
    cd comicbookcollector
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Run the collector:**
    ```bash
    go run cmd/scraper/main.go
    ```

    *Note: Currently, the target username is hardcoded in `main.go`. This will be configurable in future versions.*

### Building from Source

You can build a standalone binary for your system:

```bash
# Using standard Go build
go build -o comicbookcollector ./cmd/scraper

# Run it
./comicbookcollector
```

## 🏗️ Project Structure

This overview helps you navigate the codebase:

```text
.
├── .github/            # CI/CD Workflows (GitHub Actions)
├── cmd/
│   └── scraper/        # Application Entrypoint (main.go)
├── internal/
│   ├── browser/        # Browser initialization & generic actions
│   └── locg/           # League of Comic Geeks specific logic (Parsing, Models)
├── go.mod              # Module definition
└── README.md           # This file
```

## 📝 Roadmap

- [x] Basic Login & Session Handling
- [x] Collection Infinite Scroll
- [ ] **Detailed Metadata Extraction** (UPC, Price, Dates) (In Progress)
- [ ] SQLite Database Integration
- [ ] Config file support (YAML/JSON)
- [ ] Export formats (CSV, JSON)

## ⚠️ Disclaimer

This tool is for **personal educational purposes and backup** only. Please respect the terms of service of any website you scrape. Do not use this tool to overload servers.

---
*Maintained by [Arvid Warnecke](https://github.com/madhatter)*
