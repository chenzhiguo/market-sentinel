# Reddit 排序功能文档

## 概述

Reddit 采集器现在支持多种排序类型和时间范围，让你可以根据需求获取不同类型的内容。

## 支持的排序类型

### 1. **new** (最新)
获取最新发布的帖子，按时间倒序排列。

**适用场景：**
- 实时监控市场动态
- 获取最新的新闻和讨论
- 追踪突发事件

**示例：**
```yaml
reddit:
  enabled: true
  sort_type: "new"
  subreddits:
    - "wallstreetbets"
```

### 2. **hot** (热门)
获取当前最热门的帖子，综合考虑点赞数和时间。

**适用场景：**
- 发现当前最受关注的话题
- 了解社区热点
- 获取高质量讨论

**示例：**
```yaml
reddit:
  enabled: true
  sort_type: "hot"
  subreddits:
    - "stocks"
```

### 3. **top** (最佳)
获取指定时间范围内点赞最多的帖子。

**时间范围选项：**
- `hour` - 过去1小时
- `day` - 过去24小时
- `week` - 过去7天
- `month` - 过去30天
- `year` - 过去365天
- `all` - 所有时间

**适用场景：**
- 回顾一周/一月的重要事件
- 发现高质量内容
- 分析历史趋势

**示例：**
```yaml
reddit:
  enabled: true
  sort_type: "top"
  time_range: "week"
  subreddits:
    - "investing"
```

### 4. **rising** (上升)
获取快速获得关注的新帖子。

**适用场景：**
- 发现潜在热点
- 早期捕捉趋势
- 获取新鲜但有潜力的内容

**示例：**
```yaml
reddit:
  enabled: true
  sort_type: "rising"
  subreddits:
    - "wallstreetbets"
```

### 5. **controversial** (争议)
获取点赞和点踩数都很高的帖子。

**时间范围选项：** 同 `top`

**适用场景：**
- 发现有争议的话题
- 了解不同观点
- 分析市场分歧

**示例：**
```yaml
reddit:
  enabled: true
  sort_type: "controversial"
  time_range: "day"
  subreddits:
    - "stocks"
```

## 配置方式

### 方式一：简单配置（所有 subreddit 使用相同排序）

```yaml
collector:
  reddit:
    enabled: true
    sort_type: "hot"        # 全局排序类型
    time_range: "day"       # 全局时间范围
    subreddits:
      - "wallstreetbets"
      - "stocks"
      - "investing"
```

### 方式二：高级配置（每个 subreddit 独立配置）

```yaml
collector:
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
      
      - subreddit: "options"
        sort_type: "controversial"
        time_range: "month"
```

**注意：** 如果同时配置了 `sources` 和 `subreddits`，系统会优先使用 `sources`。

## 使用示例

### 示例 1：监控实时新闻
```yaml
reddit:
  enabled: true
  sort_type: "new"
  subreddits:
    - "wallstreetbets"
    - "stocks"
```

### 示例 2：获取每日热点
```yaml
reddit:
  enabled: true
  sort_type: "hot"
  subreddits:
    - "wallstreetbets"
    - "investing"
```

### 示例 3：周度总结
```yaml
reddit:
  enabled: true
  sort_type: "top"
  time_range: "week"
  subreddits:
    - "stocks"
    - "investing"
```

### 示例 4：混合策略
```yaml
reddit:
  enabled: true
  sources:
    # 实时监控 WSB 热点
    - subreddit: "wallstreetbets"
      sort_type: "hot"
      time_range: "day"
    
    # 获取 stocks 的周度最佳内容
    - subreddit: "stocks"
      sort_type: "top"
      time_range: "week"
    
    # 发现 investing 的上升趋势
    - subreddit: "investing"
      sort_type: "rising"
      time_range: "day"
```

## 命令行工具

### 演示程序
```bash
# 获取热门帖子
go run cmd/demo_reddit_sorts/main.go -subreddit=wallstreetbets -sort=hot

# 获取本周最佳
go run cmd/demo_reddit_sorts/main.go -subreddit=stocks -sort=top -time=week

# 获取上升趋势
go run cmd/demo_reddit_sorts/main.go -subreddit=investing -sort=rising

# 获取争议话题
go run cmd/demo_reddit_sorts/main.go -subreddit=options -sort=controversial -time=month
```

### 测试工具
```bash
# 测试不同排序类型
go run cmd/test_reddit/main.go -subreddits=wallstreetbets
```

## Source 字段格式

采集到的数据会在 `Source` 字段中标注排序类型：

- `new` 排序：`reddit:r/wallstreetbets`（默认不显示）
- 其他排序：`reddit:r/wallstreetbets:hot`、`reddit:r/stocks:top` 等

这样可以方便地追踪数据来源和排序方式。

## API 示例

### Python 客户端
```python
import requests

# 获取热门帖子的分析
response = requests.get(
    "http://localhost:8080/api/v1/news",
    params={"source": "reddit:r/wallstreetbets:hot", "limit": 10},
    headers={"Authorization": "Bearer YOUR_TOKEN"}
)

hot_news = response.json()["data"]
```

### 过滤特定排序类型
```python
# 只获取 top 排序的内容
response = requests.get(
    "http://localhost:8080/api/v1/news",
    params={"source": "reddit:r/stocks:top"},
    headers={"Authorization": "Bearer YOUR_TOKEN"}
)
```

## 最佳实践

### 1. 实时监控策略
```yaml
# 使用 new + hot 组合
sources:
  - subreddit: "wallstreetbets"
    sort_type: "new"      # 捕捉最新动态
  - subreddit: "wallstreetbets"
    sort_type: "hot"      # 关注热点话题
```

### 2. 每日摘要策略
```yaml
# 使用 top:day
sort_type: "top"
time_range: "day"
subreddits:
  - "stocks"
  - "investing"
```

### 3. 趋势发现策略
```yaml
# 使用 rising
sort_type: "rising"
subreddits:
  - "wallstreetbets"
  - "stocks"
```

### 4. 多维度分析策略
```yaml
# 同一个 subreddit 使用多种排序
sources:
  - subreddit: "wallstreetbets"
    sort_type: "hot"
  - subreddit: "wallstreetbets"
    sort_type: "top"
    time_range: "day"
  - subreddit: "wallstreetbets"
    sort_type: "rising"
```

## 性能考虑

### 采集频率
- 每个 feed 之间有 1 秒延迟，避免触发 Reddit 限流
- 建议采集间隔：15-30 分钟

### 数据量
- 每个 RSS feed 通常返回 25 条帖子
- 使用多个排序类型会成倍增加数据量

### 示例计算
```
配置：3 个 subreddit × 2 种排序 = 6 个 feed
数据量：6 × 25 = 150 条帖子/次
采集时间：6 × 1秒 + 网络时间 ≈ 10-15 秒
```

## 故障排查

### 问题 1：获取不到数据
**可能原因：**
- 排序类型拼写错误
- 时间范围无效
- subreddit 不存在

**解决方案：**
```bash
# 检查配置
go run cmd/demo_reddit_sorts/main.go -subreddit=YOUR_SUB -sort=YOUR_SORT

# 查看日志
tail -f logs/collector.log
```

### 问题 2：数据重复
**可能原因：**
- 同一个 subreddit 配置了多次
- 不同排序返回了相同的帖子

**解决方案：**
- 数据库会自动去重（基于 URL）
- 检查 `Source` 字段区分来源

### 问题 3：429 错误（限流）
**可能原因：**
- 请求过于频繁
- 配置了太多 feed

**解决方案：**
- 增加采集间隔
- 减少 subreddit 数量
- 使用更少的排序类型

## 测试

### 单元测试
```bash
go test -v ./internal/collector -run TestRedditCollector_SortTypes
go test -v ./internal/collector -run TestRedditCollector_BuildFeedURL
```

### 集成测试
```bash
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI_HotSort
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI_TopSort
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI_AdvancedConfig
```

## 更新日志

### v0.2.0 (2026-01-30)
- ✅ 添加 5 种排序类型支持（new, hot, top, rising, controversial）
- ✅ 添加时间范围支持（hour, day, week, month, year, all）
- ✅ 支持简单配置和高级配置两种模式
- ✅ Source 字段包含排序类型信息
- ✅ 修复作者格式问题
- ✅ 添加完整的测试覆盖

## 参考资料

- [Reddit RSS Feeds 文档](https://www.reddit.com/wiki/rss)
- [Reddit API 限流政策](https://github.com/reddit-archive/reddit/wiki/API)
- [项目测试文档](../internal/collector/TEST_README.md)
