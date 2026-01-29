package analyzer

import (
	"regexp"
	"strings"
)

// StockMapper maps company names and keywords to stock symbols
type StockMapper struct {
	companies map[string]string
	keywords  map[string][]string
}

func NewStockMapper() *StockMapper {
	return &StockMapper{
		companies: defaultCompanies(),
		keywords:  defaultKeywords(),
	}
}

func defaultCompanies() map[string]string {
	return map[string]string{
		// Tech Giants
		"apple":     "AAPL",
		"iphone":    "AAPL",
		"ipad":      "AAPL",
		"mac":       "AAPL",
		"tim cook":  "AAPL",
		"microsoft": "MSFT",
		"windows":   "MSFT",
		"azure":     "MSFT",
		"satya":     "MSFT",
		"google":    "GOOGL",
		"alphabet":  "GOOGL",
		"youtube":   "GOOGL",
		"amazon":    "AMZN",
		"aws":       "AMZN",
		"bezos":     "AMZN",
		"meta":      "META",
		"facebook":  "META",
		"instagram": "META",
		"zuckerberg": "META",

		// AI & Semiconductors
		"nvidia":    "NVDA",
		"jensen":    "NVDA",
		"cuda":      "NVDA",
		"amd":       "AMD",
		"lisa su":   "AMD",
		"intel":     "INTC",
		"tsmc":      "TSM",
		"broadcom":  "AVGO",
		"micron":    "MU",
		"super micro": "SMCI",
		"smci":      "SMCI",

		// EV & Auto
		"tesla":     "TSLA",
		"musk":      "TSLA",
		"spacex":    "TSLA", // Often moves together
		"byd":       "BYDDY",
		"rivian":    "RIVN",
		"lucid":     "LCID",
		"nio":       "NIO",
		"xpeng":     "XPEV",
		"li auto":   "LI",

		// Finance
		"jpmorgan":  "JPM",
		"goldman":   "GS",
		"bitcoin":   "MSTR", // Proxy
		"crypto":    "COIN", // Proxy

		// Retail
		"walmart":   "WMT",
		"costco":    "COST",
		"target":    "TGT",
		"starbucks": "SBUX",
	}
}

func defaultKeywords() map[string][]string {
	return map[string][]string{
		// Sector / Macro
		"fed":           {"SPY", "QQQ", "TLT"},
		"powell":        {"SPY", "QQQ"},
		"rate cut":      {"TLT", "IWM", "XBI"}, // Bonds, Small caps, Biotech
		"rate hike":     {"UUP", "XLF"},        // Dollar, Financials
		"inflation":     {"GLD", "TIP"},
		"cpi":           {"SPY", "QQQ"},
		"oil":           {"XOM", "CVX", "USO"},
		"opec":          {"XOM", "CVX", "USO"},
		"semiconductor": {"SOXX", "SMH"},
		"chip":          {"SOXX", "SMH"},
		"ai":            {"NVDA", "MSFT", "GOOGL"},
		"housing":       {"XHB", "ITB"},
	}
}

// FindRelatedStocks 查找文本中隐含的股票代码
func (m *StockMapper) FindRelatedStocks(text string) []string {
	text = strings.ToLower(text)
	found := make(map[string]bool)

	// 1. Check company names
	for name, symbol := range m.companies {
		if strings.Contains(text, name) {
			found[symbol] = true
		}
	}

	// 2. Check keywords (Industry/Macro)
	for keyword, symbols := range m.keywords {
		if strings.Contains(text, keyword) {
			for _, s := range symbols {
				found[s] = true
			}
		}
	}

	// 3. Check for explicit ticker mentions ($AAPL, AAPL)
	// Regex: Word boundary or $, 1-5 uppercase letters, Word boundary
	tickerRe := regexp.MustCompile(`(?:\$|\b)([A-Z]{1,5})\b`)
	matches := tickerRe.FindAllStringSubmatch(strings.ToUpper(text), -1)
	for _, match := range matches {
		if len(match) > 1 && m.isValidTicker(match[1]) {
			found[match[1]] = true
		}
	}

	var result []string
	for s := range found {
		result = append(result, s)
	}
	return result
}

func (m *StockMapper) isValidTicker(s string) bool {
	if len(s) < 1 || len(s) > 5 {
		return false
	}
	// Filter common words that look like tickers
	commonWords := map[string]bool{
		"A": true, "I": true, "AM": true, "PM": true, "AN": true,
		"THE": true, "AND": true, "FOR": true, "WITH": true, "ARE": true,
		"BUT": true, "NOT": true, "CAN": true, "ALL": true, "ANY": true,
		"NEW": true, "BIG": true, "USA": true, "CEO": true, "IPO": true,
		"GDP": true, "AI": true, "EV": true, "USD": true, "YOY": true,
		"QoQ": true, "ATH": true, "URL": true, "API": true, "APP": true,
		"NOW": true, "OUT": true, "BUY": true, "SELL": true,
	}
	return !commonWords[s]
}
