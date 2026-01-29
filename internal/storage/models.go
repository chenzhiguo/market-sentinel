package storage

import (
	"time"
)

// NewsItem represents a collected news or social media post
type NewsItem struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Source      string    `json:"source" gorm:"index:idx_news_source"`
	SourceID    string    `json:"source_id" gorm:"uniqueIndex:idx_source_source_id"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at" gorm:"index:idx_news_published"`
	CollectedAt time.Time `json:"collected_at"`
	Processed   int       `json:"processed" gorm:"index:idx_news_processed"` // 0 or 1
}

// Analysis represents AI analysis result
type Analysis struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	NewsID         string    `json:"news_id" gorm:"index:idx_analysis_news"`
	News           NewsItem  `json:"-" gorm:"foreignKey:NewsID;references:ID"`
	Sentiment      string    `json:"sentiment"`       // positive, negative, neutral
	SentimentScore float64   `json:"sentiment_score"` // -1.0 to 1.0
	Entities       []string  `json:"entities" gorm:"serializer:json"`
	RelatedStocks  []string  `json:"related_stocks" gorm:"serializer:json"`
	ImpactLevel    string    `json:"impact_level"` // high, medium, low
	Summary        string    `json:"summary"`
	KeyPoints      []string  `json:"key_points" gorm:"serializer:json"`
	AnalyzedAt     time.Time `json:"analyzed_at"`
	RawResponse    string    `json:"raw_response"`
}

// Alert represents a high-impact alert
type Alert struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	NewsID       string    `json:"news_id"`
	AnalysisID   string    `json:"analysis_id"`
	Analysis     Analysis  `json:"-" gorm:"foreignKey:AnalysisID"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Severity     string    `json:"severity"` // high, critical
	Stocks       []string  `json:"stocks" gorm:"serializer:json"`
	CreatedAt    time.Time `json:"created_at" gorm:"index:idx_alerts_created"`
	Acknowledged int       `json:"acknowledged"` // 0 or 1
}

// Report represents a generated report
type Report struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Type        string                 `json:"type"` // morning-brief, daily-summary
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Summary     string                 `json:"summary"`
	TopNews     []NewsItem             `json:"top_news" gorm:"serializer:json"`
	StockImpact map[string]StockImpact `json:"stock_impact" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at"`
	FilePath    string                 `json:"file_path"`
}

// StockImpact (Struct for JSON serialization)
type StockImpact struct {
	Symbol    string `json:"symbol"`
	Score     int    `json:"score"`
	Reasoning string `json:"reasoning"`
	Timeframe string `json:"timeframe"`
}

// StockSentiment (This is a result struct, not a table)
type StockSentiment struct {
	Symbol        string     `json:"symbol"`
	TotalMentions int        `json:"total_mentions"`
	PositiveCount int        `json:"positive_count"`
	NegativeCount int        `json:"negative_count"`
	NeutralCount  int        `json:"neutral_count"`
	OverallScore  float64    `json:"overall_score"`
	RecentNews    []NewsItem `json:"recent_news"`
	LastUpdated   time.Time  `json:"last_updated"`
}
