package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/chenzhiguo/market-sentinel/internal/collector"
	"github.com/chenzhiguo/market-sentinel/internal/config"
)

func main() {
	// Command line flags
	subreddits := flag.String("subreddits", "wallstreetbets,stocks", "Comma-separated list of subreddits")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("=== Reddit Collector Manual Test ===\n")

	// Parse subreddits
	subredditList := strings.Split(*subreddits, ",")
	for i, s := range subredditList {
		subredditList[i] = strings.TrimSpace(s)
	}

	fmt.Printf("Testing subreddits: %v\n\n", subredditList)

	// Create collector config
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: subredditList,
	}

	// Create collector
	redditCollector := collector.NewRedditCollector(cfg)

	// Collect items
	fmt.Println("Starting collection...")
	items, err := redditCollector.Collect()
	if err != nil {
		log.Fatalf("Collection failed: %v", err)
	}

	// Display results
	fmt.Printf("\n✓ Successfully collected %d items\n\n", len(items))

	// Group by source
	sourceCount := make(map[string]int)
	for _, item := range items {
		sourceCount[item.Source]++
	}

	fmt.Println("Items by source:")
	for source, count := range sourceCount {
		fmt.Printf("  - %s: %d items\n", source, count)
	}

	// Display sample items
	fmt.Println("\n--- Sample Items ---")
	displayCount := 5
	if len(items) < displayCount {
		displayCount = len(items)
	}

	for i := 0; i < displayCount; i++ {
		item := items[i]
		fmt.Printf("\n[%d] %s\n", i+1, item.Title)
		fmt.Printf("    Source: %s\n", item.Source)
		fmt.Printf("    Author: %s\n", item.Author)
		fmt.Printf("    URL: %s\n", item.URL)
		fmt.Printf("    Published: %s\n", item.PublishedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("    ID: %s\n", item.ID)

		if *verbose {
			contentPreview := item.Content
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:200] + "..."
			}
			fmt.Printf("    Content: %s\n", contentPreview)
		}
	}

	// Validation checks
	fmt.Println("\n--- Validation Checks ---")

	allValid := true

	// Check 1: All items have IDs
	missingIDs := 0
	for _, item := range items {
		if item.ID == "" {
			missingIDs++
		}
	}
	if missingIDs > 0 {
		fmt.Printf("✗ %d items missing IDs\n", missingIDs)
		allValid = false
	} else {
		fmt.Println("✓ All items have IDs")
	}

	// Check 2: All items have titles
	missingTitles := 0
	for _, item := range items {
		if item.Title == "" {
			missingTitles++
		}
	}
	if missingTitles > 0 {
		fmt.Printf("✗ %d items missing titles\n", missingTitles)
		allValid = false
	} else {
		fmt.Println("✓ All items have titles")
	}

	// Check 3: All items have valid URLs
	invalidURLs := 0
	for _, item := range items {
		if !strings.HasPrefix(item.URL, "https://www.reddit.com/") {
			invalidURLs++
		}
	}
	if invalidURLs > 0 {
		fmt.Printf("✗ %d items have invalid URLs\n", invalidURLs)
		allValid = false
	} else {
		fmt.Println("✓ All items have valid Reddit URLs")
	}

	// Check 4: All items have authors
	missingAuthors := 0
	for _, item := range items {
		if item.Author == "" || item.Author == "u/unknown" {
			missingAuthors++
		}
	}
	if missingAuthors > 0 {
		fmt.Printf("⚠ %d items have missing/unknown authors\n", missingAuthors)
	} else {
		fmt.Println("✓ All items have authors")
	}

	// Check 5: Source format
	invalidSources := 0
	for _, item := range items {
		if !strings.HasPrefix(item.Source, "reddit:r/") {
			invalidSources++
		}
	}
	if invalidSources > 0 {
		fmt.Printf("✗ %d items have invalid source format\n", invalidSources)
		allValid = false
	} else {
		fmt.Println("✓ All items have correct source format")
	}

	// Final result
	fmt.Println()
	if allValid {
		fmt.Println("✓ All validation checks passed!")
	} else {
		fmt.Println("✗ Some validation checks failed")
	}

	// Statistics
	fmt.Println("\n--- Statistics ---")

	// Average content length
	totalContentLen := 0
	for _, item := range items {
		totalContentLen += len(item.Content)
	}
	avgContentLen := 0
	if len(items) > 0 {
		avgContentLen = totalContentLen / len(items)
	}
	fmt.Printf("Average content length: %d characters\n", avgContentLen)

	// Unique authors
	uniqueAuthors := make(map[string]bool)
	for _, item := range items {
		uniqueAuthors[item.Author] = true
	}
	fmt.Printf("Unique authors: %d\n", len(uniqueAuthors))

	fmt.Println("\n=== Test Complete ===")
}
