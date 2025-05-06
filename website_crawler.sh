#!/bin/bash

# Simple Website Crawler Script
# Usage: ./website_crawler.sh [starting_url] [max_depth] [domain_restriction]

# Default values
STARTING_URL=${1:-"https://example.com"}
MAX_DEPTH=${2:-3}
DOMAIN_RESTRICTION=$(echo $STARTING_URL | awk -F/ '{print $3}')
OUTPUT_DIR="crawled_pages"
VISITED_URLS=()
CURRENT_DEPTH=1

# Check for wget dependency
if ! command -v wget &> /dev/null; then
    echo "Error: wget is required but not installed. Please install wget."
    exit 1
fi

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to sanitize URLs for filenames
sanitize_url() {
    echo "$1" | sed 's/[^a-zA-Z0-9]/_/g'
}

# Function to check if URL has already been visited
url_visited() {
    local url="$1"
    for visited in "${VISITED_URLS[@]}"; do
        if [[ "$visited" == "$url" ]]; then
            return 0
        fi
    done
    return 1
}

# Function to get base URL and directory URL
get_base_urls() {
    local url="$1"
    # Get domain (e.g., http://example.com)
    local domain=$(echo "$url" | grep -o 'https\?://[^/]*')
    
    # Get directory part (e.g., http://example.com/path/)
    local dir_url=$(echo "$url" | sed -E 's|^(https?://[^/]+/.*)/[^/]*$|\1/|')
    if [[ "$dir_url" == "$domain" || "$dir_url" == "$url" ]]; then
        dir_url="${domain}/"
    fi
    
    echo "$domain|$dir_url"
}

# Function to crawl a URL
crawl_url() {
    local url="$1"
    local depth="$2"
    
    # Check if we've reached max depth
    if [[ $depth -gt $MAX_DEPTH ]]; then
        return
    fi
    
    # Check if URL has already been visited
    if url_visited "$url"; then
        return
    fi
    
    echo "Crawling: $url (Depth: $depth)"
    
    # Add URL to visited list
    VISITED_URLS+=("$url")
    
    # Create a sanitized filename
    local filename=$(sanitize_url "$url")
    
    # Download the page
    wget -q "$url" -O "$OUTPUT_DIR/$filename.html"
    
    # If wget failed, skip this URL
    if [[ $? -ne 0 ]]; then
        echo "Failed to download $url"
        return
    fi
    
    # Get base URLs for resolving relative paths
    local base_urls=$(get_base_urls "$url")
    local domain=$(echo "$base_urls" | cut -d'|' -f1)
    local dir_url=$(echo "$base_urls" | cut -d'|' -f2)
    
    # Extract links from the page and crawl them
    local new_urls=$(grep -o 'href="[^"]*"' "$OUTPUT_DIR/$filename.html" | cut -d'"' -f2)
    
    # Process each found URL
    for new_url in $new_urls; do
        # Debug line to show original URL
        # echo "Original URL: $new_url"
        
        # Only follow links that end in .html
        if [[ "$new_url" != *".html" && "$new_url" != *".html?*" && "$new_url" != "/" ]]; then
            continue
        fi
        
        # Skip empty URLs, anchors, javascript, mailto
        if [[ -z "$new_url" || "$new_url" == "#"* || "$new_url" == "javascript:"* || "$new_url" == "mailto:"* ]]; then
            continue
        fi
        
        # Handle different types of URLs
        if [[ "$new_url" == http* ]]; then
            # Already an absolute URL, use as is
            :
        elif [[ "$new_url" == /* ]]; then
            # Absolute path within the same domain (starts with /)
            new_url="${domain}${new_url}"
        else
            # Relative URL (like "overview.html")
            new_url="${dir_url}${new_url}"
        fi
        
        # Check if URL is in the same domain (if domain restriction is enabled)
        if [[ -n "$DOMAIN_RESTRICTION" ]]; then
            if ! echo "$new_url" | grep -q "$DOMAIN_RESTRICTION"; then
                continue
            fi
        fi
        
        # Normalize URL by removing trailing slashes and fragments
        new_url=$(echo "$new_url" | sed 's/#.*$//' | sed 's|/*$||')
        
        # Debug line to show resolved URL
        # echo "Resolved URL: $new_url"
        
        # Crawl the new URL at the next depth level
        crawl_url "$new_url" $((depth + 1))
    done
}

echo "Starting crawler at $STARTING_URL with max depth $MAX_DEPTH"
echo "Restricting to domain: $DOMAIN_RESTRICTION"
echo "Output directory: $OUTPUT_DIR"
echo "Only following links that end in .html"

# Start crawling from the initial URL
crawl_url "$STARTING_URL" $CURRENT_DEPTH

echo "Crawling complete. Downloaded pages are in $OUTPUT_DIR/"