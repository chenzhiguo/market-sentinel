# Reddit 采集器快速参考

## 排序类型速查表

| 排序类型 | 说明 | 时间范围 | 使用场景 |
|---------|------|---------|---------|
| `new` | 最新帖子 | ❌ | 实时监控、突发事件 |
| `hot` | 热门帖子 | ❌ | 当前热点、高质量讨论 |
| `top` | 最佳帖子 | ✅ | 回顾总结、历史分析 |
| `rising` | 上升趋势 | ❌ | 早期发现、潜在热点 |
| `controversial` | 争议话题 | ✅ | 市场分歧、不同观点 |

## 时间范围选项

| 选项 | 说明 | 适用排序 |
|-----|------|---------|
| `hour` | 过去1小时 | top, controversial |
| `day` | 过去24小时 | top, controversial |
| `week` | 过去7天 | top, controversial |
| `month` | 过去30天 | top, controversial |
| `year` | 过去365天 | top, controversial |
| `all` | 所有时间 | top, controversial |

## 配置模板

### 实时监控
```yaml
reddit:
  enabled: true
  sort_type: "new"
  subreddits: ["wallstreetbets", "stocks"]
```

### 热点追踪
```yaml
reddit:
  enabled: true
  sort_type: "hot"
  subreddits: ["wallstreetbets", "investing"]
```

### 每日摘要
```yaml
reddit:
  enabled: true
  sort_type: "top"
  time_range: "day"
  subreddits: ["stocks", "investing"]
```

### 周度总结
```yaml
reddit:
  enabled: true
  sort_type: "top"
  time_range: "week"
  subreddits: ["stocks", "investing", "options"]
```

### 趋势发现
```yaml
reddit:
  enabled: true
  sort_type: "rising"
  subreddits: ["wallstreetbets", "stocks"]
```

### 混合策略
```yaml
reddit:
  enabled: true
  sources:
    - subreddit: "wallstreetbets"
      sort_type: "hot"
      time_range: "day"
    - subreddit: "stocks"
      sort_type: "top"
      time_range: "week"
    - subreddit: "investing"
      sort_type: "rising"
      time_range: "day"
```

## 命令行示例

```bash
# 热门帖子
go run cmd/demo_reddit_sorts/main.go -subreddit=wallstreetbets -sort=hot

# 本周最佳
go run cmd/demo_reddit_sorts/main.go -subreddit=stocks -sort=top -time=week

# 上升趋势
go run cmd/demo_reddit_sorts/main.go -subreddit=investing -sort=rising

# 争议话题
go run cmd/demo_reddit_sorts/main.go -subreddit=options -sort=controversial -time=day
```

## API 查询示例

```bash
# 获取热门帖子
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:8080/api/v1/news?source=reddit:r/wallstreetbets:hot"

# 获取本周最佳
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:8080/api/v1/news?source=reddit:r/stocks:top"

# 获取上升趋势
curl -H "Authorization: Bearer TOKEN" \
  "http://localhost:8080/api/v1/news?source=reddit:r/investing:rising"
```

## Source 字段格式

| 排序类型 | Source 格式 |
|---------|------------|
| new | `reddit:r/wallstreetbets` |
| hot | `reddit:r/wallstreetbets:hot` |
| top | `reddit:r/stocks:top` |
| rising | `reddit:r/investing:rising` |
| controversial | `reddit:r/options:controversial` |

## 常见问题

### Q: 如何同时使用多种排序？
A: 使用高级配置的 `sources` 数组，可以为同一个 subreddit 配置多个排序类型。

### Q: 时间范围对所有排序都有效吗？
A: 不是，只有 `top` 和 `controversial` 支持时间范围参数。

### Q: 默认值是什么？
A: `sort_type` 默认为 `"new"`，`time_range` 默认为 `"day"`。

### Q: 如何避免重复数据？
A: 数据库会根据 URL 自动去重，不同排序返回的相同帖子会被合并。

### Q: 采集频率建议？
A: 建议 15-30 分钟采集一次，避免触发 Reddit 限流。

## 性能参考

| 配置 | Feed 数量 | 预计数据量 | 预计时间 |
|-----|----------|-----------|---------|
| 3 个 subreddit × 1 种排序 | 3 | 75 条 | ~5 秒 |
| 3 个 subreddit × 2 种排序 | 6 | 150 条 | ~10 秒 |
| 5 个 subreddit × 1 种排序 | 5 | 125 条 | ~8 秒 |
| 5 个 subreddit × 2 种排序 | 10 | 250 条 | ~15 秒 |

## 测试命令

```bash
# 单元测试
go test -v ./internal/collector -run TestRedditCollector

# 集成测试
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI

# 特定排序测试
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI_HotSort
```

## 相关文档

- [完整功能文档](REDDIT_SORT_FEATURE.md)
- [实现总结](../REDDIT_SORT_IMPLEMENTATION.md)
- [测试文档](../internal/collector/TEST_README.md)
