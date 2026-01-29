package collector

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type RedditCollector struct {
	cfg    config.RedditConfig
	parser *gofeed.Parser
}

func NewRedditCollector(cfg config.RedditConfig) *RedditCollector {
	fp := gofeed.NewParser()
	fp.Client = &http.Client{
		Timeout: 30 * time.Second,
	}
	// gofeed 默认使用 Go-http-client/1.1，Reddit 可能会屏蔽，最好自定义 User-Agent
	// 但 gofeed 的 ParseURL 封装了 Client.Get。
	// 为了设置 Header，我们需要自定义 gofeed 的 HTTP 请求逻辑，或者简单的，我们在 fetch 时手动处理。
	
	return &RedditCollector{
		cfg:    cfg,
		parser: fp,
	}
}

func (r *RedditCollector) Collect() ([]storage.NewsItem, error) {
	var allItems []storage.NewsItem

	for _, subreddit := range r.cfg.Subreddits {
		items, err := r.fetchSubreddit(subreddit)
		if err != nil {
			fmt.Printf("Error fetching r/%s: %v\n", subreddit, err)
			continue
		}
		allItems = append(allItems, items...)
		// 礼貌性延时，避免触发限流
		time.Sleep(1 * time.Second)
	}

	return allItems, nil
}

func (r *RedditCollector) fetchSubreddit(subreddit string) ([]storage.NewsItem, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/new.rss", subreddit)

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

	feed, err := r.parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []storage.NewsItem
	for _, item := range feed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		// Reddit RSS content is HTML encoded, we might want to clean it or just store title + link
		// Title is usually enough for headlines. Content has a lot of HTML table structure.
		// For simplicity, we use Title and append Description as Content.
		
		content := item.Title
		if item.Content != "" {
			content += "\n\n" + item.Content
		} else {
			content += "\n\n" + item.Description
		}

		newsItem := storage.NewsItem{
			ID:          GenerateID("reddit", item.Link),
			Source:      "reddit:r/" + subreddit,
			Author:      "u/" + extractAuthor(item.Author),
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

func extractAuthor(author *gofeed.Person) string {
	if author != nil {
		return author.Name
	}
	return "unknown"
}
