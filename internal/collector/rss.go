package collector

import (
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type RSSCollector struct {
	cfg    config.RSSConfig
	parser *gofeed.Parser
}

func NewRSSCollector(cfg config.RSSConfig) *RSSCollector {
	return &RSSCollector{
		cfg:    cfg,
		parser: gofeed.NewParser(),
	}
}

func (r *RSSCollector) Collect() ([]storage.NewsItem, error) {
	var allItems []storage.NewsItem

	for _, feedURL := range r.cfg.Feeds {
		items, err := r.fetchFeed(feedURL)
		if err != nil {
			// Log but continue with other feeds
			continue
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

func (r *RSSCollector) fetchFeed(feedURL string) ([]storage.NewsItem, error) {
	feed, err := r.parser.ParseURL(feedURL)
	if err != nil {
		return nil, err
	}

	var items []storage.NewsItem
	for _, item := range feed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		content := item.Description
		if item.Content != "" {
			content = item.Content
		}

		newsItem := storage.NewsItem{
			ID:          GenerateID("rss", item.Link+item.Title),
			Source:      "rss:" + feed.Title,
			Author:      feed.Title,
			Content:     CleanContent(item.Title + "\n\n" + content),
			URL:         item.Link,
			PublishedAt: publishedAt,
			CollectedAt: time.Now(),
		}

		items = append(items, newsItem)
	}

	return items, nil
}
