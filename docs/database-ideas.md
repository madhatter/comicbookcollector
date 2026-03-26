# Database / Data Source Ideas

Ideas for enriching the local database beyond what LoCG provides — e.g. authors, artists, story arcs, characters.

## Metron API (primary enrichment source)

- REST API for comic book metadata, lookup by UPC possible
- Covers a broad range of publishers
- Planned integration: a reusable `EnrichByUPC(upc string)` function that fetches a defined set of metadata fields (authors, artists, story arcs, characters, …) and updates the local DB
- Use case 1: enrich existing LoCG collection after initial scrape
- Use case 2: enrich comics scanned via the barcode scanner in `../comicscanner`

## Potential data sources

### Marvel API
- Official, well-documented, reliable
- Free with API key: https://developer.marvel.com
- Covers Marvel only — good fit given the collection focus

### DC
- No public API available
- Would need a third-party source (Comic Vine or GCD)

### Image Comics
- No public API available
- Would need a third-party source

### Comic Vine
- Broad coverage across all publishers
- Free API with key: https://comicvine.gamespot.com/api/
- Previously tried but experienced frequent timeouts and ambiguous titles
- Could be revisited with proper retry logic and rate limiting

### Grand Comics Database (GCD)
- Open data, full database dump available (CC licence): https://www.comics.org
- Good fallback for publishers without an official API

## Possible approach

Use the Marvel API as the primary source for Marvel titles, and Comic Vine or GCD as a fallback for DC and Image. LoCG IDs could serve as the starting point, with additional metadata fetched and merged into the local SQLite database.

Combining multiple sources would also be a good Go learning exercise: handling retries, rate limiting, and merging data from different schemas.
