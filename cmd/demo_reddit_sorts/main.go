package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/chenzhiguo/market-sentinel/internal/collector"
	"github.com/chenzhiguo/market-sentinel/internal/config"
)

func main() {
	// Command line flags
	subreddit := flag.String("subreddit", "wallstreetbets", "Subreddit to fetch from")
	sortType := flag.String("sort", "new", "Sort type: new, hot, top, rising, controversial")
	timeRange := flag.String("time", "day", "Time range for top/controversial: hour, day, week, month, year, all")
	flag.Parse()

	fmt.Println("=== Reddit Sort Types Demo ===\n")
	fmt.Printf("Subreddit: r/%s\n", *subreddit)
	fmt.Printf("Sort: %s\n", *sortType)
	if *sortType == "top" || *sortType == "controversial" {
		fmt.Printf("Time Range: %s\n", *timeRange)
	}
	fmt.Println()

	// Create collector config
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{*subreddit},
		SortType:   *sortType,
		TimeRange:  *timeRange,
	}

	// Create collector
	redditCollector := collector.NewRedditCollector(cfg)

	// Collect items
	fmt.Println("Fetching posts...")
	items, err := redditCollector.Collect()
	if err != nil {
		log.Fatalf("Collection failed: %v", err)
	}

	fmt.Printf("\n✓ Successfully collected %d posts\n\n", len(items))

	// Display top 10 posts
	displayCount := 10
	if len(items) < displayCount {
		displayCount = len(items)
	}

	fmt.Printf("--- Top %d Posts ---\n\n", displayCount)
	for i := 0; i < displayCount; i++ {
		item := items[i]
		fmt.Printf("[%d] %s\n", i+1, item.Title)
		fmt.Printf("    Author: %s\n", item.Author)
		fmt.Printf("    URL: %s\n", item.URL)
		fmt.Printf("    Published: %s\n", item.PublishedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Source: %s\n\n", item.Source)
	}

	fmt.Println("=== Demo Complete ===")
	fmt.Println("\nTry different combinations:")
	fmt.Println("  go run cmd/demo_reddit_sorts/main.go -subreddit=stocks -sort=hot")
	fmt.Println("  go run cmd/demo_reddit_sorts/main.go -subreddit=wallstreetbets -sort=top -time=week")
	fmt.Println("  go run cmd/demo_reddit_sorts/main.go -subreddit=investing -sort=rising")
	fmt.Println("  go run cmd/demo_reddit_sorts/main.go -subreddit=options -sort=controversial -time=month")
}
