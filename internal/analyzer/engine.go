package analyzer

import (
	"log"
	"time"

	"market-sentinel/internal/storage"
)

// Engine 分析引擎
type Engine struct {
	llm         *ClaudeAnalyzer
	mapper      *StockMapper
	store       *storage.SQLiteStore
	alertChan   chan *storage.Alert
	interval    time.Duration
	stopCh      chan struct{}
}

func NewEngine(llm *ClaudeAnalyzer, mapper *StockMapper, store *storage.SQLiteStore) *Engine {
	return &Engine{
		llm:       llm,
		mapper:    mapper,
		store:     store,
		alertChan: make(chan *storage.Alert, 100),
		interval:  1 * time.Minute,
		stopCh:    make(chan struct{}),
	}
}

func (e *Engine) AlertChannel() <-chan *storage.Alert {
	return e.alertChan
}

func (e *Engine) Start() {
	go e.run()
}

func (e *Engine) Stop() {
	close(e.stopCh)
}

func (e *Engine) run() {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.processUnanalyzed()
		case <-e.stopCh:
			return
		}
	}
}

func (e *Engine) processUnanalyzed() {
	news, err := e.store.GetUnprocessedNews(10)
	if err != nil {
		log.Printf("Error getting unprocessed news: %v", err)
		return
	}

	for _, item := range news {
		e.AnalyzeNews(&item)
	}
}

// AnalyzeNews 分析单条新闻
func (e *Engine) AnalyzeNews(news *storage.NewsItem) (*storage.Analysis, error) {
	// 调用Claude分析
	analysis, err := e.llm.Analyze(news)
	if err != nil {
		log.Printf("Error analyzing news %s: %v", news.ID, err)
		return nil, err
	}

	// 补充股票映射
	if e.mapper != nil {
		additionalStocks := e.mapper.FindRelatedStocks(news.Title + " " + news.Content)
		for _, stock := range additionalStocks {
			found := false
			for _, existing := range analysis.RelatedStocks {
				if existing == stock {
					found = true
					break
				}
			}
			if !found {
				analysis.RelatedStocks = append(analysis.RelatedStocks, stock)
			}
		}
	}

	// 保存分析结果
	if err := e.store.SaveAnalysis(analysis); err != nil {
		log.Printf("Error saving analysis: %v", err)
		return nil, err
	}

	// 标记新闻已处理
	e.store.MarkNewsProcessed(news.ID)

	// 检查是否需要生成警报
	if analysis.ImpactLevel == "high" {
		alert := &storage.Alert{
			ID:          storage.NewUUID(),
			NewsID:      news.ID,
			AnalysisID:  analysis.ID,
			Title:       news.Title,
			Description: analysis.Summary,
			Severity:    "high",
			Stocks:      analysis.RelatedStocks,
			CreatedAt:   time.Now(),
		}
		e.store.SaveAlert(alert)

		select {
		case e.alertChan <- alert:
		default:
		}
	}

	return analysis, nil
}

// AnalyzeNow 立即分析所有未处理新闻
func (e *Engine) AnalyzeNow() (int, error) {
	news, err := e.store.GetUnprocessedNews(50)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range news {
		if _, err := e.AnalyzeNews(&item); err == nil {
			count++
		}
	}

	return count, nil
}
