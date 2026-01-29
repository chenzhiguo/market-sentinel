package analyzer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

// Engine manages the analysis workflow independently from collectors
type Engine struct {
	analyzer  *Analyzer
	store     *storage.Storage
	stopCh    chan struct{}
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.Mutex

	// Configuration
	pollInterval time.Duration
	workerCount  int
}

func NewEngine(analyzer *Analyzer, store *storage.Storage) *Engine {
	return &Engine{
		analyzer:     analyzer,
		store:        store,
		stopCh:       make(chan struct{}),
		pollInterval: 10 * time.Second, // 默认10秒轮询一次
		workerCount:  3,                // 默认3个并发分析
	}
}

// Start begins the background analysis loop
func (e *Engine) Start() {
	e.mu.Lock()
	if e.isRunning {
		e.mu.Unlock()
		return
	}
	e.isRunning = true
	e.stopCh = make(chan struct{})
	e.mu.Unlock()

	log.Printf("Starting Analysis Engine with %d workers...", e.workerCount)

	e.wg.Add(1)
	go e.loop()
}

// Stop gracefully shuts down the engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.isRunning {
		return
	}
	close(e.stopCh)
	e.isRunning = false
	e.wg.Wait()
	log.Println("Analysis Engine stopped")
}

func (e *Engine) loop() {
	defer e.wg.Done()
	
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.processBatch()
		}
	}
}

func (e *Engine) processBatch() {
	// 1. 获取未处理的新闻 (每次获取 workerCount * 2 条，避免频繁查询)
	batchSize := e.workerCount * 2
	items, err := e.store.GetUnprocessedNews(batchSize)
	if err != nil {
		log.Printf("Engine: failed to fetch news: %v", err)
		return
	}

	if len(items) == 0 {
		return // No work
	}

	log.Printf("Engine: processing batch of %d items", len(items))

	// 2. 使用 Worker Pool 并发处理
	var wg sync.WaitGroup
	sem := make(chan struct{}, e.workerCount) // 信号量控制并发

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{} // Acquire

		go func(news storage.NewsItem) {
			defer wg.Done()
			defer func() { <-sem }() // Release

			e.analyzeAndHandle(news)
		}(item)
	}

	wg.Wait()
}

func (e *Engine) analyzeAndHandle(item storage.NewsItem) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 分析
	analysis, err := e.analyzer.Analyze(ctx, &item)
	if err != nil {
		log.Printf("Engine: analysis failed for %s: %v", item.ID, err)
		// TODO: 可以在 DB 增加 retry_count，避免死循环失败
		return
	}

	// 保存分析结果
	if err := e.store.SaveAnalysis(analysis); err != nil {
		log.Printf("Engine: failed to save analysis %s: %v", analysis.ID, err)
		return
	}

	// 标记新闻已处理
	if err := e.store.MarkNewsProcessed(item.ID); err != nil {
		log.Printf("Engine: failed to mark processed %s: %v", item.ID, err)
		return
	}

	// 检查是否需要警报 (高影响且高置信度)
	if analysis.ImpactLevel == "high" && analysis.SentimentScore != 0 {
		e.triggerAlert(item, analysis)
	}
}

func (e *Engine) triggerAlert(news storage.NewsItem, analysis *storage.Analysis) {
	log.Printf("🚨 HIGH IMPACT ALERT: %s (Score: %.2f)", news.Title, analysis.SentimentScore)
	
	alert := &storage.Alert{
		ID:          fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		NewsID:      news.ID,
		AnalysisID:  analysis.ID,
		Title:       news.Title,
		Description: analysis.Summary,
		Severity:    "high",
		Stocks:      analysis.RelatedStocks,
		CreatedAt:   time.Now(),
	}

	if err := e.store.SaveAlert(alert); err != nil {
		log.Printf("Engine: failed to save alert: %v", err)
	}
}
