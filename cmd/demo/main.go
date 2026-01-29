package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chenzhiguo/market-sentinel/internal/analyzer"
	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化依赖
	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}
	
	// 3. 初始化分析器 (会根据配置选择 Ollama)
	ai := analyzer.New(cfg, store)

	// 4. 准备一条模拟新闻
	mockNews := &storage.NewsItem{
		ID:          "demo-news-001",
		Source:      "demo",
		Author:      "Bloomberg",
		Title:       "Apple Partners with OpenAI to Integrate ChatGPT into iOS 18",
		Content:     "Apple Inc. announced a landmark partnership with OpenAI to bring ChatGPT features to the next generation of iPhone operating system. Analysts predict this move will significantly boost iPhone upgrades cycle. Meanwhile, Google's stock fell 2% on concerns about losing search market share.",
		URL:         "https://demo.com/apple-openai",
		PublishedAt: time.Now(),
	}

	fmt.Println("------------------------------------------------")
	fmt.Printf("📰 正在分析新闻: %s\n", mockNews.Title)
	fmt.Printf("📝 内容摘要: %s\n", mockNews.Content)
	fmt.Printf("🤖 使用模型: %s (%s)\n", cfg.Analyzer.LLMProvider, cfg.Analyzer.LLMModel)
	fmt.Println("------------------------------------------------")
	fmt.Println("⏳ AI 思考中 (请求 Ollama)...")

	// 5. 执行分析
	ctx := context.Background()
	analysis, err := ai.Analyze(ctx, mockNews)
	if err != nil {
		log.Fatalf("❌ 分析失败: %v\n请检查 Ollama 是否运行 (ollama serve) 且模型已下载。", err)
	}

	// 6. 打印结果
	fmt.Println("\n✅ 分析完成！结果如下：")
	fmt.Println("------------------------------------------------")
	
	// 格式化输出
	fmt.Printf("📊 情感倾向: %s (分数: %.2f)\n", analysis.Sentiment, analysis.SentimentScore)
	fmt.Printf("💡 影响等级: %s\n", analysis.ImpactLevel)
	fmt.Printf("📝 AI 总结:  %s\n", analysis.Summary)
	fmt.Println("📈 相关股票:")
	
	// 解析 RawResponse 里的 JSON 来展示股票详情 (因为 analysis.RelatedStocks 只是字符串列表)
	// 这里简单反序列化一下 RawResponse 只是为了展示详情
	var raw struct {
		Stocks []struct {
			Symbol    string `json:"symbol"`
			Score     int    `json:"score"`
			Reasoning string `json:"reasoning"`
		} `json:"stocks"`
	}
	// 尝试从 RawResponse 提取 JSON
	json.Unmarshal([]byte(analysis.RawResponse), &raw)
	
	if len(raw.Stocks) > 0 {
		for _, stock := range raw.Stocks {
			scoreIcon := "➖"
			if stock.Score > 0 { scoreIcon = "🟢" }
			if stock.Score < 0 { scoreIcon = "🔴" }
			fmt.Printf("   %s %-5s: %d 分 | %s\n", scoreIcon, stock.Symbol, stock.Score, stock.Reasoning)
		}
	} else {
		// Fallback if parsing failed
		fmt.Printf("   %v\n", analysis.RelatedStocks)
	}

	fmt.Println("------------------------------------------------")
}
