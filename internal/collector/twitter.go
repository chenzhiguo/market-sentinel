package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type TwitterCollector struct {
	cfg    config.TwitterConfig
	client *http.Client
}

func NewTwitterCollector(cfg config.TwitterConfig) *TwitterCollector {
	return &TwitterCollector{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *TwitterCollector) Collect() ([]storage.NewsItem, error) {
	var allItems []storage.NewsItem

	for _, account := range t.cfg.Accounts {
		items, err := t.fetchAccount(account)
		if err != nil {
			// Log but continue with other accounts
			fmt.Printf("Error fetching @%s: %v\n", account, err)
			continue
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

func (t *TwitterCollector) fetchAccount(account string) ([]storage.NewsItem, error) {
	// Try Nitter instances
	for _, host := range t.cfg.NitterHosts {
		items, err := t.fetchFromNitter(host, account)
		if err == nil {
			return items, nil
		}
	}

	return nil, fmt.Errorf("all Nitter instances failed for @%s", account)
}

func (t *TwitterCollector) fetchFromNitter(host, account string) ([]storage.NewsItem, error) {
	// Fetch RSS feed from Nitter
	url := fmt.Sprintf("%s/%s/rss", host, account)
	resp, err := t.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nitter returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return t.parseNitterRSS(string(body), account, host)
}

func (t *TwitterCollector) parseNitterRSS(xmlContent, account, host string) ([]storage.NewsItem, error) {
	var items []storage.NewsItem

	// Simple regex-based parsing for Nitter RSS
	// In production, use proper XML parsing
	itemRe := regexp.MustCompile(`<item>(.*?)</item>`)
	titleRe := regexp.MustCompile(`<title><!\[CDATA\[(.*?)\]\]></title>`)
	linkRe := regexp.MustCompile(`<link>(.*?)</link>`)
	pubDateRe := regexp.MustCompile(`<pubDate>(.*?)</pubDate>`)

	matches := itemRe.FindAllStringSubmatch(xmlContent, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		itemContent := match[1]

		var content, link, pubDate string

		if m := titleRe.FindStringSubmatch(itemContent); len(m) > 1 {
			content = m[1]
		}
		if m := linkRe.FindStringSubmatch(itemContent); len(m) > 1 {
			link = m[1]
			// Convert Nitter link back to Twitter
			link = strings.Replace(link, host, "https://twitter.com", 1)
		}
		if m := pubDateRe.FindStringSubmatch(itemContent); len(m) > 1 {
			pubDate = m[1]
		}

		if content == "" {
			continue
		}

		publishedAt, _ := time.Parse(time.RFC1123Z, pubDate)

		items = append(items, storage.NewsItem{
			ID:          GenerateID("twitter", content),
			Source:      "twitter",
			Author:      "@" + account,
			AuthorID:    account,
			Content:     CleanContent(content),
			URL:         link,
			PublishedAt: publishedAt,
			CollectedAt: time.Now(),
		})
	}

	return items, nil
}

// FetchWithAPI fetches tweets using Twitter API v2 (requires API key)
func (t *TwitterCollector) FetchWithAPI(account, bearerToken string) ([]storage.NewsItem, error) {
	// Get user ID first
	userURL := fmt.Sprintf("https://api.twitter.com/2/users/by/username/%s", account)
	req, _ := http.NewRequest("GET", userURL, nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, err
	}

	// Fetch tweets
	tweetsURL := fmt.Sprintf("https://api.twitter.com/2/users/%s/tweets?max_results=10&tweet.fields=created_at,text", userResp.Data.ID)
	req, _ = http.NewRequest("GET", tweetsURL, nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err = t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tweetsResp struct {
		Data []struct {
			ID        string `json:"id"`
			Text      string `json:"text"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tweetsResp); err != nil {
		return nil, err
	}

	var items []storage.NewsItem
	for _, tweet := range tweetsResp.Data {
		publishedAt, _ := time.Parse(time.RFC3339, tweet.CreatedAt)
		items = append(items, storage.NewsItem{
			ID:          GenerateID("twitter", tweet.ID),
			Source:      "twitter",
			Author:      "@" + account,
			AuthorID:    account,
			Content:     CleanContent(tweet.Text),
			URL:         fmt.Sprintf("https://twitter.com/%s/status/%s", account, tweet.ID),
			PublishedAt: publishedAt,
			CollectedAt: time.Now(),
		})
	}

	return items, nil
}
