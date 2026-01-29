package collector

import (
	"time"

	"market-sentinel/internal/storage"
)

// Collector 数据采集器接口
type Collector interface {
	Name() string
	Collect() ([]storage.NewsItem, error)
}

// Manager 采集管理器
type Manager struct {
	collectors []Collector
	store      *storage.SQLiteStore
	interval   time.Duration
	stopCh     chan struct{}
}

func NewManager(store *storage.SQLiteStore, interval time.Duration) *Manager {
	return &Manager{
		store:    store,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (m *Manager) Register(c Collector) {
	m.collectors = append(m.collectors, c)
}

func (m *Manager) Start() {
	go m.run()
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

func (m *Manager) run() {
	// 立即执行一次
	m.collectAll()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectAll()
		case <-m.stopCh:
			return
		}
	}
}

func (m *Manager) collectAll() {
	for _, c := range m.collectors {
		items, err := c.Collect()
		if err != nil {
			continue
		}

		for _, item := range items {
			m.store.SaveNews(&item)
		}
	}
}

// CollectNow 立即执行一次采集
func (m *Manager) CollectNow() (int, error) {
	total := 0
	for _, c := range m.collectors {
		items, err := c.Collect()
		if err != nil {
			continue
		}

		for _, item := range items {
			if err := m.store.SaveNews(&item); err == nil {
				total++
			}
		}
	}
	return total, nil
}
