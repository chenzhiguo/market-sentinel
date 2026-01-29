package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Collector CollectorConfig `mapstructure:"collector"`
	Analyzer  AnalyzerConfig  `mapstructure:"analyzer"`
	Reporter  ReporterConfig  `mapstructure:"reporter"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type AuthConfig struct {
	Tokens    []string `mapstructure:"tokens"`
	RateLimit int      `mapstructure:"rate_limit"` // requests per minute
}

type StorageConfig struct {
	Database   string `mapstructure:"database"`
	ReportsDir string `mapstructure:"reports_dir"`
}

type CollectorConfig struct {
	ScanInterval time.Duration `mapstructure:"scan_interval"`
	Sources      []string      `mapstructure:"sources"`
	Twitter      TwitterConfig `mapstructure:"twitter"`
	Reddit       RedditConfig  `mapstructure:"reddit"`
	RSS          RSSConfig     `mapstructure:"rss"`
}

type TwitterConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	NitterHosts  []string `mapstructure:"nitter_hosts"`
	Accounts     []string `mapstructure:"accounts"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
}

type RedditConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	Subreddits   []string `mapstructure:"subreddits"` // e.g., ["wallstreetbets", "investing"]
}

type RSSConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Feeds   []string `mapstructure:"feeds"`
}

type AnalyzerConfig struct {
	LLMProvider string `mapstructure:"llm_provider"` // anthropic, ollama
	LLMModel    string `mapstructure:"llm_model"`
	APIKey      string `mapstructure:"api_key"`
	OllamaURL   string `mapstructure:"ollama_url"` // e.g. "http://localhost:11434"
}

type ReporterConfig struct {
	SaveToFile bool   `mapstructure:"save_to_file"`
	FileFormat string `mapstructure:"file_format"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("auth.rate_limit", 100)
	v.SetDefault("storage.database", "./data/sentinel.db")
	v.SetDefault("storage.reports_dir", "./data/reports")
	v.SetDefault("collector.scan_interval", "15m")
	v.SetDefault("collector.sources", []string{"twitter", "rss"})
	v.SetDefault("analyzer.llm_provider", "anthropic")
	v.SetDefault("analyzer.llm_model", "claude-sonnet-4-20250514")
	v.SetDefault("reporter.save_to_file", true)
	v.SetDefault("reporter.file_format", "json")

	// Config file
	v.SetConfigFile(path)

	// Environment variables
	v.SetEnvPrefix("SENTINEL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Override with environment variables for sensitive data
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.Analyzer.APIKey = apiKey
	}
	if token := os.Getenv("SENTINEL_API_TOKEN"); token != "" {
		cfg.Auth.Tokens = append(cfg.Auth.Tokens, token)
	}

	return &cfg, nil
}
