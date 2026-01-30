package collector

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type RedditCollector struct {
	cfg    config.RedditConfig
	parser *gofeed.Parser
}

// RedditFeedConfig represents configuration for a single Reddit feed
type RedditFeedConfig struct {
	Subreddit string
	SortType  string // new, hot, top, rising, controversial
	TimeRange string // hour, day, week, month, year, all
}

func NewRedditCollector(cfg config.RedditConfig) *RedditCollector {
	fp := gofeed.NewParser()
	fp.Client = &http.Client{
		Timeout: 30 * time.Second,
	}

	return &RedditCollector{
		cfg:    cfg,
		parser: fp,
	}
}

func (r *RedditCollector) Collect() ([]storage.NewsItem, error) {
	var allItems []storage.NewsItem

	// Build feed configurations
	feeds := r.buildFeedConfigs()

	for _, feed := range feeds {
		items, err := r.fetchFeed(feed)
		if err != nil {
			fmt.Printf("Error fetching r/%s (%s): %v\n", feed.Subreddit, feed.SortType, err)
			continue
		}
		allItems = append(allItems, items...)
		// 礼貌性延时，避免触发限流
		time.Sleep(1 * time.Second)
	}

	return allItems, nil
}

// buildFeedConfigs creates feed configurations from config
func (r *RedditCollector) buildFeedConfigs() []RedditFeedConfig {
	var feeds []RedditFeedConfig

	// If advanced sources are configured, use them
	if len(r.cfg.Sources) > 0 {
		for _, source := range r.cfg.Sources {
			feeds = append(feeds, RedditFeedConfig{
				Subreddit: source.Subreddit,
				SortType:  r.normalizeSortType(source.SortType),
				TimeRange: r.normalizeTimeRange(source.TimeRange),
			})
		}
		return feeds
	}

	// Otherwise, use simple subreddit list with global sort settings
	defaultSort := r.normalizeSortType(r.cfg.SortType)
	defaultTime := r.normalizeTimeRange(r.cfg.TimeRange)

	for _, subreddit := range r.cfg.Subreddits {
		feeds = append(feeds, RedditFeedConfig{
			Subreddit: subreddit,
			SortType:  defaultSort,
			TimeRange: defaultTime,
		})
	}

	return feeds
}

// normalizeSortType ensures valid sort type
func (r *RedditCollector) normalizeSortType(sortType string) string {
	sortType = strings.ToLower(strings.TrimSpace(sortType))
	validSorts := map[string]bool{
		"new": true, "hot": true, "top": true,
		"rising": true, "controversial": true,
	}
	if validSorts[sortType] {
		return sortType
	}
	return "new" // default
}

// normalizeTimeRange ensures valid time range
func (r *RedditCollector) normalizeTimeRange(timeRange string) string {
	timeRange = strings.ToLower(strings.TrimSpace(timeRange))
	validRanges := map[string]bool{
		"hour": true, "day": true, "week": true,
		"month": true, "year": true, "all": true,
	}
	if validRanges[timeRange] {
		return timeRange
	}
	return "day" // default
}

// fetchFeed fetches a single Reddit feed with specified parameters
func (r *RedditCollector) fetchFeed(feed RedditFeedConfig) ([]storage.NewsItem, error) {
	url := r.buildFeedURL(feed)

	// 手动创建请求以设置 User-Agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "market-sentinel/0.1.0")

	resp, err := r.parser.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit returned status %d", resp.StatusCode)
	}

	parsedFeed, err := r.parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []storage.NewsItem
	for _, item := range parsedFeed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		content := item.Title
		if item.Content != "" {
			content += "\n\n" + item.Content
		} else {
			content += "\n\n" + item.Description
		}

		// Include sort type in source for tracking
		source := fmt.Sprintf("reddit:r/%s", feed.Subreddit)
		if feed.SortType != "new" {
			source = fmt.Sprintf("reddit:r/%s:%s", feed.Subreddit, feed.SortType)
		}

		newsItem := storage.NewsItem{
			ID:          GenerateID("reddit", item.Link),
			Source:      source,
			Author:      extractAuthor(item.Author),
			Title:       item.Title,
			Content:     CleanContent(content),
			URL:         item.Link,
			PublishedAt: publishedAt,
			CollectedAt: time.Now(),
		}

		items = append(items, newsItem)
	}

	return items, nil
}

// buildFeedURL constructs the Reddit RSS URL with sort and time parameters
func (r *RedditCollector) buildFeedURL(feed RedditFeedConfig) string {
	// Base URL format:
	// new:    https://www.reddit.com/r/{subreddit}/new.rss
	// hot:    https://www.reddit.com/r/{subreddit}/hot.rss
	// top:    https://www.reddit.com/r/{subreddit}/top.rss?t=day
	// rising: https://www.reddit.com/r/{subreddit}/rising.rss
	// controversial: https://www.reddit.com/r/{subreddit}/controversial.rss?t=day

	baseURL := fmt.Sprintf("https://www.reddit.com/r/%s/%s.rss", feed.Subreddit, feed.SortType)

	// Add time range for top and controversial
	if feed.SortType == "top" || feed.SortType == "controversial" {
		baseURL += "?t=" + feed.TimeRange
	}

	return baseURL
}

// fetchSubreddit is kept for backward compatibility
func (r *RedditCollector) fetchSubreddit(subreddit string) ([]storage.NewsItem, error) {
	feed := RedditFeedConfig{
		Subreddit: subreddit,
		SortType:  r.normalizeSortType(r.cfg.SortType),
		TimeRange: r.normalizeTimeRange(r.cfg.TimeRange),
	}
	return r.fetchFeed(feed)
}

func extractAuthor(author *gofeed.Person) string {
	if author != nil && author.Name != "" {
		// Reddit RSS author format is "/u/username"
		// Remove leading "/" if present
		name := author.Name
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		// If it doesn't start with "u/", add it
		if !strings.HasPrefix(name, "u/") {
			name = "u/" + name
		}
		return name
	}
	return "u/unknown"
}
