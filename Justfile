# Justfile

# Default recipe
default: build

# Build the application
build:
    @echo "Building binary..."
    go build -o bin/comicbookcollector ./cmd/scraper

# Run the scraper
run:
    go run ./cmd/scraper/main.go

# Run tests
test:
    go test -v ./...

# Clean up build artifacts and locks
clean:
    rm -rf bin/
    rm -f chrome-data/SingletonLock
    @echo "Cleaned up."

# Install dependencies
deps:
    go mod download
    go mod tidy

## Metron API exploration commands

explore_dir := "docs/api-responses"

# Ensure output directory exists
_setup:
    @mkdir -p {{explore_dir}}

# 1. Validate credentials (1 request)
@explore-validate: _setup
    echo "Validating credentials..."
    http -a madhatter:$METRON_PASS \
        https://metron.cloud/api/publisher/?page_size=1 \
        > {{explore_dir}}/validate.json
    echo "✅ Credentials valid"

# 2. Find issue by UPC
@explore-upc UPC: _setup
    echo "Searching for UPC: {{UPC}}"
    http -a madhatter:$METRON_PASS \
        https://metron.cloud/api/issue/ \
        upc=={{UPC}} \
        > {{explore_dir}}/upc-{{UPC}}.json
    jq -r '.results[] | "Found: \(.series.name) #\(.number) (ID: \(.id))"' \
        {{explore_dir}}/upc-{{UPC}}.json

# 3. Get full issue details
@explore-issue ID: _setup
    echo "Getting details for issue {{ID}}"
    http -a madhatter:$METRON_PASS \
        https://metron.cloud/api/issue/{{ID}}/ \
        > {{explore_dir}}/issue-{{ID}}.json
    echo "✅ Saved to {{explore_dir}}/issue-{{ID}}.json"
    echo -e "\nCredits:"
    jq -r '.credits[] | "  \(.creator.name): \(.role | map(.name) | join(", "))"' \
        {{explore_dir}}/issue-{{ID}}.json

# 4. Complete exploration workflow
@explore-full UPC: _setup
    echo "=== Full exploration for UPC: {{UPC}} ==="
    just explore-upc {{UPC}}
    $(eval ISSUE_ID := $(shell jq -r '.results[0].id' {{explore_dir}}/upc-{{UPC}}.json))
    just explore-issue $ISSUE_ID

# 5. List all saved responses
@explore-list:
    @echo "Saved API responses:"
    @ls -lh {{explore_dir}}/*.json 2>/dev/null || echo "No responses yet"

# 6. View a response with syntax highlighting
@explore-view FILE:
    @jq '.' {{explore_dir}}/{{FILE}}

# 7. Compare two responses
@explore-diff FILE1 FILE2:
    @diff -u \
        <(jq -S '.' {{explore_dir}}/{{FILE1}}) \
        <(jq -S '.' {{explore_dir}}/{{FILE2}})
