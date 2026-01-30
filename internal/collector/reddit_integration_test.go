//go:build integration
// +build integration

package collector

import (
	"strings"
	"testing"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/config"
)

// TestRedditCollector_RealAPI_Wallstreetbets tests real Reddit RSS feed
// Run with: go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI
func TestRedditCollector_RealAPI_Wallstreetbets(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"wallstreetbets"},
	}

	collector := NewRedditCollector(cfg)

	items, err := collector.fetchSubreddit("wallstreetbets")
	if err != nil {
		t.Fatalf("Failed to fetch r/wallstreetbets: %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected at least one item from r/wallstreetbets")
	}

	// Validate first item structure
	if len(items) > 0 {
		item := items[0]

		if item.ID == "" {
			t.Error("Expected item to have an ID")
		}

		if item.Source != "reddit:r/wallstreetbets" {
			t.Errorf("Expected source 'reddit:r/wallstreetbets', got '%s'", item.Source)
		}

		if item.Title == "" {
			t.Error("Expected item to have a title")
		}

		if item.URL == "" {
			t.Error("Expected item to have a URL")
		}

		if !strings.HasPrefix(item.URL, "https://www.reddit.com/r/") {
			t.Errorf("Expected Reddit URL, got '%s'", item.URL)
		}

		if item.Author == "" {
			t.Error("Expected item to have an author")
		}

		if item.PublishedAt.IsZero() {
			t.Error("Expected item to have a published time")
		}

		if item.CollectedAt.IsZero() {
			t.Error("Expected item to have a collected time")
		}

		t.Logf("Sample item: Title='%s', Author='%s', URL='%s'",
			item.Title, item.Author, item.URL)
	}
}

// TestRedditCollector_RealAPI_MultipleSubreddits tests multiple subreddits
func TestRedditCollector_RealAPI_MultipleSubreddits(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"wallstreetbets", "stocks", "investing"},
	}

	collector := NewRedditCollector(cfg)

	start := time.Now()
	items, err := collector.Collect()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to collect from subreddits: %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected at least one item from all subreddits")
	}

	// Should have items from multiple sources
	sources := make(map[string]int)
	for _, item := range items {
		sources[item.Source]++
	}

	t.Logf("Collected %d items from %d sources in %v", len(items), len(sources), elapsed)

	for source, count := range sources {
		t.Logf("  %s: %d items", source, count)
	}

	// Verify rate limiting (should take at least 2 seconds for 3 subreddits)
	if elapsed < 2*time.Second {
		t.Logf("Warning: Collection took %v, expected at least 2s for rate limiting", elapsed)
	}
}

// TestRedditCollector_RealAPI_InvalidSubreddit tests error handling
func TestRedditCollector_RealAPI_InvalidSubreddit(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"thissubredditdoesnotexist123456789"},
	}

	collector := NewRedditCollector(cfg)

	items, err := collector.fetchSubreddit("thissubredditdoesnotexist123456789")

	// Reddit might return 404 or empty feed
	if err != nil {
		t.Logf("Expected error for invalid subreddit: %v", err)
	}

	if len(items) > 0 {
		t.Errorf("Expected no items from invalid subreddit, got %d", len(items))
	}
}

// TestRedditCollector_RealAPI_ContentQuality tests content extraction quality
func TestRedditCollector_RealAPI_ContentQuality(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"stocks"},
	}

	collector := NewRedditCollector(cfg)

	items, err := collector.fetchSubreddit("stocks")
	if err != nil {
		t.Fatalf("Failed to fetch r/stocks: %v", err)
	}

	if len(items) == 0 {
		t.Fatal("Expected at least one item from r/stocks")
	}

	// Check content quality
	for i, item := range items {
		if i >= 3 {
			break // Check first 3 items
		}

		t.Logf("\n--- Item %d ---", i+1)
		t.Logf("Title: %s", item.Title)
		t.Logf("Author: %s", item.Author)
		t.Logf("URL: %s", item.URL)
		t.Logf("Published: %s", item.PublishedAt.Format(time.RFC3339))
		t.Logf("Content length: %d chars", len(item.Content))

		// Validate content is not just title
		if item.Content == item.Title {
			t.Logf("Warning: Content is same as title (might be missing description)")
		}

		// Check for HTML artifacts (should be cleaned)
		if len(item.Content) > 0 {
			if item.Content[0:1] == "<" {
				t.Errorf("Content appears to contain HTML: %s", item.Content[:50])
			}
		}
	}
}

// TestRedditCollector_RealAPI_UserAgent tests that User-Agent is properly set
func TestRedditCollector_RealAPI_UserAgent(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"wallstreetbets"},
	}

	collector := NewRedditCollector(cfg)

	// This test verifies that requests don't get blocked by Reddit
	// Reddit blocks default Go user agents
	items, err := collector.fetchSubreddit("wallstreetbets")

	if err != nil {
		// If we get 403 or 429, User-Agent might be the issue
		t.Fatalf("Failed to fetch (possible User-Agent issue): %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected items, got none (possible User-Agent blocking)")
	} else {
		t.Logf("Successfully fetched %d items with custom User-Agent", len(items))
	}
}

// TestRedditCollector_RealAPI_Timing tests collection timing
func TestRedditCollector_RealAPI_Timing(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"wallstreetbets"},
	}

	collector := NewRedditCollector(cfg)

	start := time.Now()
	items, err := collector.fetchSubreddit("wallstreetbets")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}

	t.Logf("Fetched %d items in %v", len(items), elapsed)

	// Should complete within reasonable time (< 10s)
	if elapsed > 10*time.Second {
		t.Errorf("Fetch took too long: %v", elapsed)
	}

	// Should not be instant (network request)
	if elapsed < 100*time.Millisecond {
		t.Errorf("Fetch was suspiciously fast: %v", elapsed)
	}
}

// TestRedditCollector_RealAPI_HotSort tests hot sort
func TestRedditCollector_RealAPI_HotSort(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"wallstreetbets"},
		SortType:   "hot",
	}

	collector := NewRedditCollector(cfg)
	items, err := collector.Collect()

	if err != nil {
		t.Fatalf("Failed to collect hot posts: %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected at least one hot post")
	}

	// Verify source includes sort type
	if len(items) > 0 {
		if !strings.Contains(items[0].Source, ":hot") {
			t.Errorf("Expected source to contain ':hot', got '%s'", items[0].Source)
		}
		t.Logf("Hot post: %s", items[0].Title)
	}
}

// TestRedditCollector_RealAPI_TopSort tests top sort with time range
func TestRedditCollector_RealAPI_TopSort(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"stocks"},
		SortType:   "top",
		TimeRange:  "week",
	}

	collector := NewRedditCollector(cfg)
	items, err := collector.Collect()

	if err != nil {
		t.Fatalf("Failed to collect top posts: %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected at least one top post")
	}

	if len(items) > 0 {
		if !strings.Contains(items[0].Source, ":top") {
			t.Errorf("Expected source to contain ':top', got '%s'", items[0].Source)
		}
		t.Logf("Top post of the week: %s", items[0].Title)
	}
}

// TestRedditCollector_RealAPI_RisingSort tests rising sort
func TestRedditCollector_RealAPI_RisingSort(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled:    true,
		Subreddits: []string{"investing"},
		SortType:   "rising",
	}

	collector := NewRedditCollector(cfg)
	items, err := collector.Collect()

	if err != nil {
		t.Fatalf("Failed to collect rising posts: %v", err)
	}

	// Rising might have fewer posts
	t.Logf("Collected %d rising posts", len(items))

	if len(items) > 0 {
		if !strings.Contains(items[0].Source, ":rising") {
			t.Errorf("Expected source to contain ':rising', got '%s'", items[0].Source)
		}
		t.Logf("Rising post: %s", items[0].Title)
	}
}

// TestRedditCollector_RealAPI_AdvancedConfig tests per-subreddit configuration
func TestRedditCollector_RealAPI_AdvancedConfig(t *testing.T) {
	cfg := config.RedditConfig{
		Enabled: true,
		Sources: []config.RedditSource{
			{
				Subreddit: "wallstreetbets",
				SortType:  "hot",
				TimeRange: "day",
			},
			{
				Subreddit: "stocks",
				SortType:  "top",
				TimeRange: "week",
			},
		},
	}

	collector := NewRedditCollector(cfg)
	items, err := collector.Collect()

	if err != nil {
		t.Fatalf("Failed to collect with advanced config: %v", err)
	}

	if len(items) == 0 {
		t.Error("Expected at least one item")
	}

	// Group by source
	sources := make(map[string]int)
	for _, item := range items {
		sources[item.Source]++
	}

	t.Logf("Collected %d items from %d sources", len(items), len(sources))
	for source, count := range sources {
		t.Logf("  %s: %d items", source, count)
	}

	// Verify we have both sources
	hasHot := false
	hasTop := false
	for source := range sources {
		if strings.Contains(source, ":hot") {
			hasHot = true
		}
		if strings.Contains(source, ":top") {
			hasTop = true
		}
	}

	if !hasHot {
		t.Error("Expected to find hot posts")
	}
	if !hasTop {
		t.Error("Expected to find top posts")
	}
}

// TestRedditCollector_RealAPI_TopTimeRanges tests different time ranges for top
func TestRedditCollector_RealAPI_TopTimeRanges(t *testing.T) {
	timeRanges := []string{"day", "week", "month"}

	for _, timeRange := range timeRanges {
		t.Run("TimeRange_"+timeRange, func(t *testing.T) {
			cfg := config.RedditConfig{
				Enabled:    true,
				Subreddits: []string{"wallstreetbets"},
				SortType:   "top",
				TimeRange:  timeRange,
			}

			collector := NewRedditCollector(cfg)
			items, err := collector.Collect()

			if err != nil {
				t.Fatalf("Failed to collect top posts for %s: %v", timeRange, err)
			}

			t.Logf("Collected %d top posts for time range '%s'", len(items), timeRange)

			if len(items) > 0 {
				t.Logf("Sample: %s", items[0].Title)
			}
		})
	}
}
