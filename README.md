# Market Sentinel 🛰️

量化舆情分析系统 - 实时监控金融新闻和社交媒体，AI 分析市场影响。

## 功能特性

- 📡 **数据采集**: Twitter/X（通过 Nitter）、RSS 财经快讯
- 🤖 **AI 分析**: 使用 Claude 进行情感分析、实体提取、股票关联
- 📊 **报告生成**: 即时警报 + 定时汇总
- 🌐 **HTTP API**: RESTful API，支持 Token 认证
- 💾 **双写存储**: SQLite 数据库 + 本地文件

## 快速开始

### 环境要求

- Go 1.22+
- SQLite3
- Anthropic API Key

### 本地运行

```bash
# 克隆项目
cd market-sentinel

# 安装依赖
go mod tidy

# 设置环境变量
export ANTHROPIC_API_KEY=sk-ant-xxxxx
export SENTINEL_API_TOKEN=your-secure-token

# 编译
go build -o sentinel ./cmd/sentinel

# 启动服务
./sentinel serve
```

### Docker 部署

```bash
# 设置环境变量
export ANTHROPIC_API_KEY=sk-ant-xxxxx
export SENTINEL_API_TOKEN=your-secure-token

# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## API 文档

### 认证

所有 API（除 `/health`）需要认证：

```bash
# Header 方式
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/reports

# Query 参数方式（调试用）
curl http://localhost:8080/api/v1/reports?token=YOUR_TOKEN
```

### 端点列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查（无需认证） |
| GET | `/api/v1/news` | 新闻列表 |
| GET | `/api/v1/news/:id` | 新闻详情 |
| GET | `/api/v1/analysis` | 分析结果列表 |
| GET | `/api/v1/analysis/:id` | 分析详情 |
| GET | `/api/v1/reports` | 报告列表 |
| GET | `/api/v1/reports/latest` | 最新报告 |
| GET | `/api/v1/reports/:id` | 报告详情 |
| GET | `/api/v1/stocks/:symbol/sentiment` | 股票舆情评分 |
| GET | `/api/v1/alerts` | 高影响事件警报 |
| POST | `/api/v1/scan` | 手动触发扫描 |

### 通用查询参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `since` | ISO8601 | 起始时间 |
| `until` | ISO8601 | 结束时间 |
| `limit` | int | 返回数量（默认 50，最大 200） |
| `offset` | int | 分页偏移 |
| `source` | string | 数据源过滤 |
| `impact` | string | 影响级别过滤（high/medium/low） |

### 响应格式

```json
// 成功
{
  "success": true,
  "data": { ... },
  "meta": {
    "total": 100,
    "limit": 50,
    "offset": 0
  }
}

// 错误
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing token"
  }
}
```

## 命令行用法

```bash
# 启动 API 服务 + 定时采集
./sentinel serve

# 单次扫描
./sentinel scan --once

# 生成报告
./sentinel report --type morning-brief
./sentinel report --type daily-summary

# 查看版本
./sentinel version
```

## 配置说明

主配置文件: `configs/config.yaml`

```yaml
server:
  host: "0.0.0.0"
  port: 8080

auth:
  tokens:
    - "your-api-token"  # 或使用 SENTINEL_API_TOKEN 环境变量

collector:
  scan_interval: 15m
  twitter:
    enabled: true
    accounts:
      - "DeItaone"
      - "elonmusk"
  rss:
    enabled: true
    feeds:
      - "https://feeds.bloomberg.com/markets/news.rss"

analyzer:
  llm_provider: "anthropic"
  llm_model: "claude-sonnet-4-20250514"
  # API key: ANTHROPIC_API_KEY 环境变量
```

## 与量化系统集成

Python 客户端示例：

```python
import requests

class SentinelClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url.rstrip('/')
        self.headers = {"Authorization": f"Bearer {token}"}
    
    def get_stock_sentiment(self, symbol: str) -> dict:
        resp = requests.get(
            f"{self.base_url}/api/v1/stocks/{symbol}/sentiment",
            headers=self.headers
        )
        return resp.json()["data"]
    
    def get_latest_report(self) -> dict:
        resp = requests.get(
            f"{self.base_url}/api/v1/reports/latest",
            headers=self.headers
        )
        return resp.json()["data"]

# 使用
client = SentinelClient("http://your-server:8080", "your-token")
nvda_sentiment = client.get_stock_sentiment("NVDA")
print(f"NVDA sentiment: {nvda_sentiment['sentiment_score']}")
```

## 目录结构

```
market-sentinel/
├── cmd/sentinel/          # 入口
├── internal/
│   ├── api/              # HTTP API
│   ├── collector/        # 数据采集
│   ├── analyzer/         # AI 分析
│   ├── reporter/         # 报告生成
│   ├── storage/          # 数据存储
│   └── config/           # 配置管理
├── configs/              # 配置文件
├── data/                 # 数据目录
│   ├── sentinel.db       # SQLite 数据库
│   └── reports/          # 报告文件
├── Dockerfile
├── docker-compose.yaml
└── README.md
```

## License

MIT
