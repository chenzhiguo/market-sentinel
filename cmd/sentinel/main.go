package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chenzhiguo/market-sentinel/internal/analyzer"
	"github.com/chenzhiguo/market-sentinel/internal/api"
	"github.com/chenzhiguo/market-sentinel/internal/collector"
	"github.com/chenzhiguo/market-sentinel/internal/config"
	"github.com/chenzhiguo/market-sentinel/internal/storage"
)

var (
	version   = "0.1.0"
	buildTime = "unknown"
)

func main() {
	// Commands
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	// Serve flags
	serveConfigPath := serveCmd.String("config", "configs/config.yaml", "Path to config file")

	// Scan flags
	scanOnce := scanCmd.Bool("once", false, "Run scan once and exit")
	scanConfigPath := scanCmd.String("config", "configs/config.yaml", "Path to config file")

	// Report flags
	reportType := reportCmd.String("type", "summary", "Report type: summary, morning-brief, alerts")
	reportConfigPath := reportCmd.String("config", "configs/config.yaml", "Path to config file")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd.Parse(os.Args[2:])
		runServe(*serveConfigPath)

	case "scan":
		scanCmd.Parse(os.Args[2:])
		runScan(*scanConfigPath, *scanOnce)

	case "report":
		reportCmd.Parse(os.Args[2:])
		runReport(*reportConfigPath, *reportType)

	case "version":
		versionCmd.Parse(os.Args[2:])
		fmt.Printf("Market Sentinel v%s (built: %s)\n", version, buildTime)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Market Sentinel - 量化舆情分析系统

Usage:
  sentinel <command> [options]

Commands:
  serve     Start the API server and collector
  scan      Run news/social media scan
  report    Generate reports
  version   Show version info

Examples:
  sentinel serve --config configs/config.yaml
  sentinel scan --once
  sentinel report --type morning-brief

Use "sentinel <command> --help" for more information.`)
}

func runServe(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// 1. Start Collector Manager (Producers)
	colManager := collector.NewManager(cfg, store)
	colManager.Start()
	defer colManager.Stop()

	// 2. Start Analysis Engine (Consumers)
	ai := analyzer.New(cfg, store)
	engine := analyzer.NewEngine(ai, store)
	engine.Start()
	defer engine.Stop()

	// 3. Start API Server
	server := api.NewServer(cfg, store)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	log.Printf("Market Sentinel started on %s:%d", cfg.Server.Host, cfg.Server.Port)
	<-done
	log.Println("Shutting down...")
}

func runScan(configPath string, once bool) {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// 1. Run Collectors
	colManager := collector.NewManager(cfg, store)
	log.Println("Running collectors...")
	colManager.RunOnce() // Synchronous run

	// 2. Trigger Analysis (if requested)
	// Even if run once, we might want to process what we just collected
	if once {
		log.Println("Running analysis on new items...")
		ai := analyzer.New(cfg, store)
		
		// Use manual batch processing instead of full Engine for CLI tool simplicity
		unprocessed, _ := store.GetUnprocessedNews(20)
		ctx := context.Background()
		for _, item := range unprocessed {
			log.Printf("Analyzing: %s", item.Title)
			ai.AnalyzeAndSave(ctx, &item)
			store.MarkNewsProcessed(item.ID)
		}
		log.Println("Scan completed")
	} else {
		// If not once, maybe loop? But typical CLI scan is one-off.
		// If user wants loop, they should use 'serve'
	}
}

func runReport(configPath string, reportType string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	log.Printf("Generating %s report...", reportType)
	// TODO: Implement report generation
}
