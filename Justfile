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
metron_user := "madhatter"
metron_pass := env_var('METRON_PASS')
explore_dir := "docs/api-responses"
rate_limit_file := env_var('HOME') + "/.metron-last-request"

# Rate-limited GET request (body only)
_metron-get endpoint:
    #!/usr/bin/env bash
    set -euo pipefail
    
    # Rate limiting: ensure 4s between requests
    if [ -f "{{rate_limit_file}}" ]; then
        ELAPSED=$(($(date +%s) - $(cat {{rate_limit_file}})))
        if [ $ELAPSED -lt 4 ]; then
            WAIT=$((4 - ELAPSED))
            echo "⏱  Rate limiting: waiting ${WAIT}s..." >&2
            sleep $WAIT
        fi
    fi
    
    # Make request
    xh GET "https://metron.cloud/api{{endpoint}}" \
        -a "{{metron_user}}:{{metron_pass}}"
    
    # Track request time
    date +%s > "{{rate_limit_file}}"

# Get headers only (for rate limit checking)
_metron-headers endpoint:
    #!/usr/bin/env bash
    set -euo pipefail
    
    # Rate limiting
    if [ -f "{{rate_limit_file}}" ]; then
        ELAPSED=$(($(date +%s) - $(cat {{rate_limit_file}})))
        if [ $ELAPSED -lt 4 ]; then
            WAIT=$((4 - ELAPSED))
            echo "⏱  Rate limiting: waiting ${WAIT}s..." >&2
            sleep $WAIT
        fi
    fi
    
    # Get headers only
    xh -h GET "https://metron.cloud/api{{endpoint}}" \
        -a "{{metron_user}}:{{metron_pass}}"
    
    # Track request time
    date +%s > "{{rate_limit_file}}"

# Ensure output directory exists
_setup:
    @mkdir -p {{explore_dir}}

# Validate credentials (1 request)
metron-validate: _setup
    #!/usr/bin/env bash
    echo "Validating credentials..."
    RESPONSE=$(just _metron-get "/publisher/?page_size=1")
    COUNT=$(echo "$RESPONSE" | jq -r '.count')
    if [ "$COUNT" -gt 0 ]; then
        echo "✅ Credentials valid ($COUNT publishers found)"
    else
        echo "❌ Unexpected response"
        exit 1
    fi

# Find issue by UPC
metron-upc UPC: _setup
    #!/usr/bin/env bash
    echo "Searching for UPC: {{UPC}}"
    RESPONSE=$(just _metron-get "/issue/?upc={{UPC}}")
    echo "$RESPONSE" > "{{explore_dir}}/upc-{{UPC}}.json"
    echo "📄 Saved to {{explore_dir}}/upc-{{UPC}}.json"
    
    # Show results
    echo "$RESPONSE" | jq -r '.results[] | "Found: \(.series.name) #\(.number) (ID: \(.id))"'

# Get issue details
metron-issue ID: _setup
    #!/usr/bin/env bash
    echo "Getting details for issue {{ID}}..."
    RESPONSE=$(just _metron-get "/issue/{{ID}}/")
    echo "$RESPONSE" > "{{explore_dir}}/issue-{{ID}}.json"
    echo "📄 Saved to {{explore_dir}}/issue-{{ID}}.json"
    
    # Show credits
    echo ""
    echo "Credits:"
    echo "$RESPONSE" | jq -r '.credits[] | " (\(.id)) \(.creator): \(.role | map(.name) | join(", "))"'

# Get creator details
metron-creator ID: _setup
    #!/usr/bin/env bash
    echo "Getting creator {{ID}}..."
    RESPONSE=$(just _metron-get "/creator/{{ID}}/")
    echo "$RESPONSE" > "{{explore_dir}}/creator-{{ID}}.json"
    echo "📄 Saved to {{explore_dir}}/creator-{{ID}}.json"
    
    # Show info
    echo "$RESPONSE" | jq -r '"Name: \(.name)\nBirth: \(.birth // "unknown")\nDeath: \(.death // "alive")"'

# Get series details
metron-series ID: _setup
    #!/usr/bin/env bash
    echo "Getting series {{ID}}..."
    RESPONSE=$(just _metron-get "/series/{{ID}}/")
    echo "$RESPONSE" > "{{explore_dir}}/series-{{ID}}.json"
    echo "📄 Saved to {{explore_dir}}/series-{{ID}}.json"
    
    # Show info
    echo "$RESPONSE" | jq -r '"Series: \(.name) v\(.volume)\nYear: \(.year_began)-\(.year_end // "ongoing")\nIssues: \(.issue_count)"'

# Show current rate limit status
metron-status:
    #!/usr/bin/env bash
    echo "Checking rate limits..."
    HEADERS=$(just _metron-headers "/publisher/?page_size=1")
    
    echo ""
    echo "Rate Limits:"
    echo "$HEADERS" | grep -i "x-ratelimit-burst" | sed 's/^/  /'
    echo "$HEADERS" | grep -i "x-ratelimit-sustained" | sed 's/^/  /'

# Complete exploration for a UPC
metron-explore UPC: _setup
    #!/usr/bin/env bash
    echo "=== Full exploration for UPC: {{UPC}} ==="
    echo ""
    
    # Step 1: Find issue
    just metron-upc {{UPC}}
    
    # Step 2: Get issue ID
    ISSUE_ID=$(jq -r '.results[0].id' "{{explore_dir}}/upc-{{UPC}}.json")
    
    if [ "$ISSUE_ID" = "null" ] || [ -z "$ISSUE_ID" ]; then
        echo "❌ No issue found for UPC {{UPC}}"
        exit 1
    fi
    
    echo ""
    
    # Step 3: Get full issue details
    just metron-issue $ISSUE_ID
    
    echo ""
    echo "✅ Exploration complete"

# List all saved responses
metron-list:
    @echo "Saved API responses:"
    @ls -lh {{explore_dir}}/*.json 2>/dev/null || echo "No responses saved yet"

# View a saved response
metron-view FILE:
    @jq '.' {{explore_dir}}/{{FILE}}

# Clean up saved responses
metron-clean:
    @rm -rf {{explore_dir}}/*.json
    @echo "✅ Cleaned up {{explore_dir}}"
