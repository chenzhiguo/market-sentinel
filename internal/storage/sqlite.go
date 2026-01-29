package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Storage struct {
	db         *gorm.DB
	backupPath string
}

// New creates a new storage instance using GORM
func New(dbPath string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.AutoMigrate(&NewsItem{}, &Analysis{}, &Alert{}, &Report{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	// Ignore WAL error
	sqlDB.Exec("PRAGMA journal_mode=WAL;")

	return &Storage{
		db:         db,
		backupPath: "",
	}, nil
}

// SaveNews 保存新闻
func (s *Storage) SaveNews(news *NewsItem) error {
	result := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source"}, {Name: "source_id"}},
		DoNothing: true,
	}).Create(news)

	if result.Error != nil {
		return result.Error
	}

	if s.backupPath != "" {
		return s.backupToFile("news", news.ID, news)
	}
	return nil
}

// ListNews 获取新闻列表（支持过滤和分页）
func (s *Storage) ListNews(since, until time.Time, source string, limit, offset int) ([]NewsItem, int, error) {
	var items []NewsItem
	var total int64

	tx := s.db.Model(&NewsItem{})

	if !since.IsZero() {
		tx = tx.Where("published_at >= ?", since)
	}
	if !until.IsZero() {
		tx = tx.Where("published_at <= ?", until)
	}
	if source != "" {
		tx = tx.Where("source = ?", source)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("published_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

// GetNews 获取单条新闻
func (s *Storage) GetNews(id string) (*NewsItem, error) {
	var item NewsItem
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetUnprocessedNews 获取未处理的新闻
func (s *Storage) GetUnprocessedNews(limit int) ([]NewsItem, error) {
	var items []NewsItem
	err := s.db.Where("processed = ?", 0).
		Order("published_at DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

// MarkNewsProcessed 标记新闻为已处理
func (s *Storage) MarkNewsProcessed(newsID string) error {
	return s.db.Model(&NewsItem{}).
		Where("id = ?", newsID).
		Update("processed", 1).Error
}

// SaveAnalysis 保存分析结果
func (s *Storage) SaveAnalysis(analysis *Analysis) error {
	if err := s.db.Create(analysis).Error; err != nil {
		return err
	}
	if s.backupPath != "" {
		return s.backupToFile("analysis", analysis.ID, analysis)
	}
	return nil
}

// ListAnalysis 获取分析列表
func (s *Storage) ListAnalysis(since, until time.Time, impact string, limit, offset int) ([]Analysis, int, error) {
	var items []Analysis
	var total int64

	tx := s.db.Model(&Analysis{})

	if !since.IsZero() {
		tx = tx.Where("analyzed_at >= ?", since)
	}
	if !until.IsZero() {
		tx = tx.Where("analyzed_at <= ?", until)
	}
	if impact != "" {
		tx = tx.Where("impact_level = ?", impact)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("analyzed_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

// GetAnalysis 获取单条分析
func (s *Storage) GetAnalysis(id string) (*Analysis, error) {
	var item Analysis
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// SaveAlert 保存警报
func (s *Storage) SaveAlert(alert *Alert) error {
	if err := s.db.Create(alert).Error; err != nil {
		return err
	}
	if s.backupPath != "" {
		return s.backupToFile("alerts", alert.ID, alert)
	}
	return nil
}

// ListAlerts 获取警报列表
func (s *Storage) ListAlerts(severity string, limit, offset int) ([]Alert, int, error) {
	var items []Alert
	var total int64

	tx := s.db.Model(&Alert{})

	if severity != "" {
		tx = tx.Where("severity = ?", severity)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

// SaveReport 保存报告
func (s *Storage) SaveReport(report *Report) error {
	if err := s.db.Create(report).Error; err != nil {
		return err
	}
	if s.backupPath != "" {
		return s.backupToFile("reports", report.ID, report)
	}
	return nil
}

// ListReports 获取报告列表
func (s *Storage) ListReports(reportType string, limit, offset int) ([]Report, int, error) {
	var items []Report
	var total int64

	tx := s.db.Model(&Report{})

	if reportType != "" {
		tx = tx.Where("type = ?", reportType)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

// GetReport 获取报告详情
func (s *Storage) GetReport(id string) (*Report, error) {
	var item Report
	if err := s.db.First(&item, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetLatestReport 获取最新报告
func (s *Storage) GetLatestReport() (*Report, error) {
	var item Report
	err := s.db.Order("created_at DESC").First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// GetStockSentiment 获取股票舆情评分
func (s *Storage) GetStockSentiment(symbol string, hours int) (*StockSentiment, error) {
	var analyses []Analysis
	// 默认 24 小时，如果传入 hours 则使用传入值
	if hours <= 0 {
		hours = 24
	}
	cutoff := time.Now().Add(time.Duration(-hours) * time.Hour)

	// 使用 LIKE 查询 JSON 数组字符串
	err := s.db.Where("related_stocks LIKE ? AND analyzed_at > ?", "%"+symbol+"%", cutoff).
		Order("analyzed_at DESC").
		Preload("News").
		Find(&analyses).Error

	if err != nil {
		return nil, err
	}

	sentiment := &StockSentiment{
		Symbol:      symbol,
		LastUpdated: time.Now(),
	}

	var totalScore float64
	for _, a := range analyses {
		totalScore += a.SentimentScore
		sentiment.TotalMentions++

		switch a.Sentiment {
		case "positive":
			sentiment.PositiveCount++
		case "negative":
			sentiment.NegativeCount++
		default:
			sentiment.NeutralCount++
		}

		if len(sentiment.RecentNews) < 5 {
			if a.News.ID != "" {
				sentiment.RecentNews = append(sentiment.RecentNews, a.News)
			}
		}
	}

	if sentiment.TotalMentions > 0 {
		sentiment.OverallScore = totalScore / float64(sentiment.TotalMentions)
	}

	return sentiment, nil
}

func (s *Storage) backupToFile(category, id string, data interface{}) error {
	dir := filepath.Join(s.backupPath, category, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	filePath := filepath.Join(dir, id+".json")
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, content, 0644)
}

func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
