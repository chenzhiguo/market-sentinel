package collector

import (
	"log"
	"sync"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

// SubCollector defines the interface for individual source collectors
type SubCollector interface {
	Collect() ([]storage.NewsItem, error)
}

// Manager orchestrates all collectors
type Manager struct {
	cfg        *config.Config
	store      *storage.Storage
	collectors []SubCollector
	stopCh     chan struct{}
	wg         sync.WaitGroup
	isRunning  bool
	mu         sync.Mutex
}

func NewManager(cfg *config.Config, store *storage.Storage) *Manager {
	m := &Manager{
		cfg:    cfg,
		store:  store,
		stopCh: make(chan struct{}),
	}

	// Initialize enabled collectors
	if cfg.Collector.Twitter.Enabled {
		m.collectors = append(m.collectors, NewTwitterCollector(cfg.Collector.Twitter))
	}
	if cfg.Collector.RSS.Enabled {
		m.collectors = append(m.collectors, NewRSSCollector(cfg.Collector.RSS))
	}
	if cfg.Collector.Reddit.Enabled {
		m.collectors = append(m.collectors, NewRedditCollector(cfg.Collector.Reddit))
	}

	return m
}

// Start begins the collection loop in background
func (m *Manager) Start() {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = true
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	log.Printf("Starting Collector Manager with %d sources...", len(m.collectors))
	
	m.wg.Add(1)
	go m.loop()
}

// Stop gracefully shuts down collectors
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.isRunning {
		return
	}
	close(m.stopCh)
	m.isRunning = false
	m.wg.Wait()
	log.Println("Collector Manager stopped")
}

func (m *Manager) loop() {
	defer m.wg.Done()

	// Initial run
	m.RunOnce()

	ticker := time.NewTicker(m.cfg.Collector.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.RunOnce()
		}
	}
}

// RunOnce executes all collectors concurrently and saves results
func (m *Manager) RunOnce() {
	var wg sync.WaitGroup
	
	log.Println("Collector: starting scan cycle...")

	for _, col := range m.collectors {
		wg.Add(1)
		go func(c SubCollector) {
			defer wg.Done()
			
			items, err := c.Collect()
			if err != nil {
				log.Printf("Collector error: %v", err)
				return
			}

			count := 0
			for _, item := range items {
				if err := m.store.SaveNews(&item); err == nil {
					count++
				}
			}
			if len(items) > 0 {
				log.Printf("Collected %d items from source", len(items))
			}
		}(col)
	}

	wg.Wait()
	log.Println("Collector: scan cycle completed")
}
