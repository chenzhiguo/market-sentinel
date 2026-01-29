package collector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type Collector struct {
	cfg   *config.Config
	store *storage.Storage

	twitter *TwitterCollector
	rss     *RSSCollector

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

func New(cfg *config.Config, store *storage.Storage) *Collector {
	c := &Collector{
		cfg:   cfg,
		store: store,
	}

	if cfg.Collector.Twitter.Enabled {
		c.twitter = NewTwitterCollector(cfg.Collector.Twitter)
	}
	if cfg.Collector.RSS.Enabled {
		c.rss = NewRSSCollector(cfg.Collector.RSS)
	}

	return c
}

func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("collector already running")
	}
	c.running = true
	c.stopCh = make(chan struct{})
	c.mu.Unlock()

	ticker := time.NewTicker(c.cfg.Collector.ScanInterval)
	defer ticker.Stop()

	// Run immediately on start
	c.scan()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopCh:
			return nil
		case <-ticker.C:
			c.scan()
		}
	}
}

func (c *Collector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		close(c.stopCh)
		c.running = false
	}
}

func (c *Collector) scan() {
	log.Println("Starting scan...")
	var wg sync.WaitGroup

	// Collect from Twitter/X
	if c.twitter != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := c.twitter.Collect()
			if err != nil {
				log.Printf("Twitter collection error: %v", err)
				return
			}
			for _, item := range items {
				if err := c.store.SaveNews(&item); err != nil {
					log.Printf("Failed to save news: %v", err)
				}
			}
			log.Printf("Collected %d items from Twitter", len(items))
		}()
	}

	// Collect from RSS
	if c.rss != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := c.rss.Collect()
			if err != nil {
				log.Printf("RSS collection error: %v", err)
				return
			}
			for _, item := range items {
				if err := c.store.SaveNews(&item); err != nil {
					log.Printf("Failed to save news: %v", err)
				}
			}
			log.Printf("Collected %d items from RSS", len(items))
		}()
	}

	wg.Wait()
	log.Println("Scan completed")
}

func (c *Collector) RunOnce() error {
	c.scan()
	return nil
}

// GenerateID creates a unique ID for a news item
func GenerateID(source, content string) string {
	h := sha256.New()
	h.Write([]byte(source))
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// CleanContent removes extra whitespace and normalizes text
func CleanContent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	return s
}
