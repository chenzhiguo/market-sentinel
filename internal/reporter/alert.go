package reporter

import (
	"fmt"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

type AlertManager struct {
	store *storage.Storage
}

func NewAlertManager(store *storage.Storage) *AlertManager {
	return &AlertManager{store: store}
}

// CheckAndCreateAlert evaluates an analysis and creates an alert if needed
func (am *AlertManager) CheckAndCreateAlert(analysis *storage.Analysis, news *storage.NewsItem) (*storage.Alert, error) {
	// Only create alerts for high impact items
	if analysis.Impact != "high" {
		return nil, nil
	}

	// Build alert title
	title := buildAlertTitle(analysis, news)

	// Build alert message
	message := buildAlertMessage(analysis, news)

	// Extract stock symbols
	var stocks []string
	for _, s := range analysis.Stocks {
		stocks = append(stocks, s.Symbol)
	}

	alert := &storage.Alert{
		ID:         fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		NewsID:     analysis.NewsID,
		AnalysisID: analysis.ID,
		Level:      determineAlertLevel(analysis),
		Title:      title,
		Message:    message,
		Stocks:     stocks,
		CreatedAt:  time.Now(),
		Notified:   false,
	}

	if err := am.store.SaveAlert(alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func buildAlertTitle(analysis *storage.Analysis, news *storage.NewsItem) string {
	sentiment := "📰"
	switch analysis.Sentiment {
	case "positive":
		sentiment = "📈"
	case "negative":
		sentiment = "📉"
	}

	source := "Unknown"
	if news != nil {
		source = news.Author
	}

	return fmt.Sprintf("%s 高影响事件 | %s", sentiment, source)
}

func buildAlertMessage(analysis *storage.Analysis, news *storage.NewsItem) string {
	msg := analysis.Summary

	if len(analysis.Stocks) > 0 {
		msg += "\n\n关联股票: "
		for i, s := range analysis.Stocks {
			if i > 0 {
				msg += ", "
			}
			scoreSign := "+"
			if s.Score < 0 {
				scoreSign = ""
			}
			msg += fmt.Sprintf("%s (%s%d)", s.Symbol, scoreSign, s.Score)
		}
	}

	return msg
}

func determineAlertLevel(analysis *storage.Analysis) string {
	// Critical if high confidence and strong sentiment
	if analysis.Confidence > 0.8 {
		for _, s := range analysis.Stocks {
			if s.Score >= 8 || s.Score <= -8 {
				return "critical"
			}
		}
	}
	return "high"
}

// GetPendingAlerts returns alerts that haven't been notified yet
func (am *AlertManager) GetPendingAlerts() ([]storage.Alert, error) {
	alerts, _, err := am.store.ListAlerts("", 100, 0)
	if err != nil {
		return nil, err
	}

	var pending []storage.Alert
	for _, a := range alerts {
		if !a.Notified {
			pending = append(pending, a)
		}
	}
	return pending, nil
}
