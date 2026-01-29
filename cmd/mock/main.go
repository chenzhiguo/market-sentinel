package main

import (
	"log"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

func main() {
	store, err := storage.New("./data/sentinel.db")
	if err != nil {
		log.Fatal(err)
	}

	// 模拟一条 Elon Musk 的推文
	tweet := &storage.NewsItem{
		ID:          "mock-tweet-001",
		Source:      "twitter",
		Author:      "@elonmusk",
		Content:     "We are making significant progress on FSD v12. It will be mind-blowing. Tesla is an AI/Robotics company, not just a car company.",
		URL:         "https://twitter.com/elonmusk/status/mock1",
		PublishedAt: time.Now(),
		CollectedAt: time.Now(),
		Processed:   0,
	}

	if err := store.SaveNews(tweet); err != nil {
		log.Fatal(err)
	}

	log.Println("Mock tweet inserted")
}
