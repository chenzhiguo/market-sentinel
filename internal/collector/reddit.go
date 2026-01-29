package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"market-sentinel/internal/storage"
)

// RedditCollector Reddit采集器
type RedditCollector struct {
	subreddits []string
	client     *http.Client
}

func NewRedditCollector(subreddits []string) *RedditCollector {
	return &RedditCollector{
		subreddits: subreddits,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *RedditCollector) Name() string {
	return "reddit"
}

func (c *RedditCollector) Collect() ([]storage.NewsItem, error) {
	var items []storage.NewsItem

	for _, subreddit := range c.subreddits {
		posts, err := c.fetchSubreddit(subreddit)
		if err != nil {
			continue
		}
		items = append(items, posts...)
	}

	return items, nil
}

func (c *RedditCollector) fetchSubreddit(subreddit string) ([]storage.NewsItem, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=25", subreddit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "MarketSentinel/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Children []struct {
				Data struct {
					ID        string  `json:"id"`
					Title     string  `json:"title"`
					Selftext  string  `json:"selftext"`
					Author    string  `json:"author"`
					Permalink string  `json:"permalink"`
					Created   float64 `json:"created_utc"`
					Score     int     `json:"score"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var items []storage.NewsItem
	for _, child := range result.Data.Children {
		post := child.Data
		item := storage.NewsItem{
			ID:          storage.NewUUID(),
			Source:      "reddit",
			SourceID:    post.ID,
			Author:      fmt.Sprintf("r/%s u/%s", subreddit, post.Author),
			Title:       post.Title,
			Content:     post.Selftext,
			URL:         "https://reddit.com" + post.Permalink,
			PublishedAt: time.Unix(int64(post.Created), 0),
			CollectedAt: time.Now(),
		}
		items = append(items, item)
	}

	return items, nil
}
