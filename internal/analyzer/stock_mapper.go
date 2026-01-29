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
		"macbook":   "AAPL",
		"tim cook":  "AAPL",
		"microsoft": "MSFT",
		"windows":   "MSFT",
		"azure":     "MSFT",
		"satya nadella": "MSFT",
		"google":    "GOOGL",
		"alphabet":  "GOOGL",
		"sundar":    "GOOGL",
		"amazon":    "AMZN",
		"aws":       "AMZN",
		"bezos":     "AMZN",
		"andy jassy": "AMZN",
		"meta":      "META",
		"facebook":  "META",
		"instagram": "META",
		"whatsapp":  "META",
		"zuckerberg": "META",

		// AI & Semiconductors
		"nvidia":    "NVDA",
		"jensen huang": "NVDA",
		"cuda":      "NVDA",
		"amd":       "AMD",
		"lisa su":   "AMD",
		"intel":     "INTC",
		"qualcomm":  "QCOM",
		"broadcom":  "AVGO",
		"tsmc":      "TSM",
		"arm":       "ARM",
		"asml":      "ASML",

		// EV & Auto
		"tesla":     "TSLA",
		"elon musk": "TSLA",
		"musk":      "TSLA",
		"spacex":    "TSLA",
		"rivian":    "RIVN",
		"lucid":     "LCID",
		"nio":       "NIO",
		"xpeng":     "XPEV",
		"li auto":   "LI",
		"ford":      "F",
		"gm":        "GM",

		// China Tech
		"alibaba":   "BABA",
		"jack ma":   "BABA",
		"tencent":   "TCEHY",
		"baidu":     "BIDU",
		"jd.com":    "JD",
		"pinduoduo": "PDD",
		"bytedance": "PDD", // Related play
		"tiktok":    "META", // Competitor

		// Finance
		"jpmorgan":  "JPM",
		"goldman":   "GS",
		"blackrock": "BLK",
		"berkshire": "BRK.B",
		"buffett":   "BRK.B",
		"visa":      "V",
		"mastercard": "MA",
		"paypal":    "PYPL",
		"square":    "SQ",
		"block":     "SQ",

		// Crypto Related
		"coinbase":  "COIN",
		"microstrategy": "MSTR",
		"bitcoin":   "MSTR",
		"btc":       "MSTR",

		// Healthcare
		"pfizer":    "PFE",
		"moderna":   "MRNA",
		"johnson":   "JNJ",
		"unitedhealth": "UNH",

		// Energy
		"exxon":     "XOM",
		"chevron":   "CVX",
		"shell":     "SHEL",
		"bp":        "BP",

		// Retail
		"walmart":   "WMT",
		"costco":    "COST",
		"target":    "TGT",
		"nike":      "NKE",
		"starbucks": "SBUX",
	}
}

func defaultKeywords() map[string][]string {
	return map[string][]string{
		// Sector keywords
		"tariff":      {"BABA", "PDD", "NIO", "XPEV"},
		"tariffs":     {"BABA", "PDD", "NIO", "XPEV"},
		"china trade": {"BABA", "PDD", "AAPL", "TSLA"},
		"chip ban":    {"NVDA", "AMD", "INTC", "TSM", "ASML"},
		"semiconductor": {"NVDA", "AMD", "INTC", "TSM", "ASML"},
		"ai regulation": {"NVDA", "GOOGL", "MSFT", "META"},
		"antitrust":   {"GOOGL", "META", "AAPL", "AMZN"},
		"rate hike":   {"JPM", "GS", "BAC", "XLF"},
		"rate cut":    {"AAPL", "MSFT", "NVDA", "QQQ"},
		"fed":         {"SPY", "QQQ", "TLT"},
		"powell":      {"SPY", "QQQ", "TLT", "JPM"},
		"yellen":      {"SPY", "TLT"},
		"oil":         {"XOM", "CVX", "OXY"},
		"opec":        {"XOM", "CVX", "OXY"},
		"ev":          {"TSLA", "RIVN", "NIO", "LCID"},
		"electric vehicle": {"TSLA", "RIVN", "NIO", "LCID"},
	}
}

func (m *StockMapper) FindStocks(text string) []string {
	text = strings.ToLower(text)
	found := make(map[string]bool)

	// Check company names
	for name, symbol := range m.companies {
		if strings.Contains(text, name) {
			found[symbol] = true
		}
	}

	// Check keywords
	for keyword, symbols := range m.keywords {
		if strings.Contains(text, keyword) {
			for _, s := range symbols {
				found[s] = true
			}
		}
	}

	// Check for explicit ticker mentions ($AAPL, AAPL)
	tickerRe := regexp.MustCompile(`\$?([A-Z]{1,5})`)
	matches := tickerRe.FindAllStringSubmatch(strings.ToUpper(text), -1)
	for _, match := range matches {
		if len(match) > 1 && isValidTicker(match[1]) {
			found[match[1]] = true
		}
	}

	var result []string
	for s := range found {
		result = append(result, s)
	}
	return result
}

func isValidTicker(s string) bool {
	// Basic validation - in production, check against a real list
	if len(s) < 1 || len(s) > 5 {
		return false
	}
	// Filter common words
	commonWords := map[string]bool{
		"A": true, "I": true, "AM": true, "PM": true, "AN": true,
		"THE": true, "AND": true, "FOR": true, "WITH": true,
	}
	return !commonWords[s]
}
