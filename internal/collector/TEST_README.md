# Reddit Collector Tests

这个目录包含 Reddit 采集器的测试用例。

## 测试文件

### 1. `reddit_test.go` - 单元测试
测试 Reddit 采集器的基本功能，不需要真实的网络请求。

**测试用例：**
- `TestRedditCollector_NewCollector` - 测试采集器初始化
- `TestRedditCollector_ExtractAuthor` - 测试作者提取逻辑
- `TestRedditCollector_ContentFormatting` - 测试内容格式化
- `TestRedditCollector_Timeout` - 测试超时配置
- `TestRedditCollector_Collect_MultipleSubreddits` - 测试多个 subreddit 配置

**运行单元测试：**
```bash
go test -v ./internal/collector -run TestRedditCollector
```

### 2. `reddit_integration_test.go` - 集成测试
测试真实的 Reddit RSS API，需要网络连接。

**测试用例：**
- `TestRedditCollector_RealAPI_Wallstreetbets` - 测试从 r/wallstreetbets 采集
- `TestRedditCollector_RealAPI_MultipleSubreddits` - 测试多个 subreddit 采集
- `TestRedditCollector_RealAPI_InvalidSubreddit` - 测试错误处理
- `TestRedditCollector_RealAPI_ContentQuality` - 测试内容质量
- `TestRedditCollector_RealAPI_UserAgent` - 测试 User-Agent 设置
- `TestRedditCollector_RealAPI_Timing` - 测试性能

**运行集成测试：**
```bash
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI
```

## 快速开始

### 运行所有单元测试
```bash
cd /path/to/market-sentinel
go test -v ./internal/collector
```

### 运行集成测试（需要网络）
```bash
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI
```

### 运行特定测试
```bash
# 只测试初始化
go test -v ./internal/collector -run TestRedditCollector_NewCollector

# 只测试真实 API
go test -tags=integration -v ./internal/collector -run TestRedditCollector_RealAPI_Wallstreetbets
```

## 测试覆盖率

生成覆盖率报告：
```bash
go test -coverprofile=coverage.out ./internal/collector
go tool cover -html=coverage.out -o coverage.html
```

## 已知问题和限制

### 1. Reddit RSS 限制
- Reddit RSS 每个 feed 通常只返回最近 25 条帖子
- 可能会遇到速率限制（429 错误）
- 某些 subreddit 可能需要认证

### 2. User-Agent 要求
Reddit 会屏蔽默认的 Go HTTP 客户端 User-Agent。当前实现使用：
```
User-Agent: market-sentinel/0.1.0
```

如果遇到 403 错误，可能需要更新 User-Agent。

### 3. 内容格式
Reddit RSS 返回的内容可能包含 HTML。当前实现：
- 使用 `gofeed` 库自动解析
- 通过 `CleanContent()` 函数清理
- 可能需要进一步的 HTML 清理

### 4. 时间解析
- Reddit RSS 使用 RFC3339 格式
- 如果解析失败，使用当前时间

## 测试结果示例

### 成功的测试输出
```
=== RUN   TestRedditCollector_RealAPI_Wallstreetbets
    reddit_integration_test.go:45: Sample item: Title='NVDA earnings beat!', Author='u/testuser', URL='https://www.reddit.com/r/wallstreetbets/...'
--- PASS: TestRedditCollector_RealAPI_Wallstreetbets (1.23s)
```

### 失败的测试（网络问题）
```
=== RUN   TestRedditCollector_RealAPI_Wallstreetbets
    reddit_integration_test.go:25: Failed to fetch r/wallstreetbets: Get "https://www.reddit.com/r/wallstreetbets/new.rss": dial tcp: lookup www.reddit.com: no such host
--- FAIL: TestRedditCollector_RealAPI_Wallstreetbets (0.05s)
```

## 改进建议

### 短期改进
1. **Mock 服务器测试** - 为单元测试添加完整的 mock HTTP 服务器
2. **错误重试** - 添加自动重试机制
3. **更好的 HTML 清理** - 改进内容清理逻辑

### 长期改进
1. **Reddit API 集成** - 使用官方 Reddit API 而不是 RSS
2. **认证支持** - 支持 OAuth 认证以访问私有 subreddit
3. **增量采集** - 只采集新内容，避免重复
4. **缓存机制** - 缓存已采集的内容

## 故障排查

### 测试失败：timeout
```
Error: context deadline exceeded
```
**解决方案：** 检查网络连接，或增加超时时间

### 测试失败：403 Forbidden
```
Error: reddit returned status 403
```
**解决方案：** 
1. 检查 User-Agent 设置
2. 可能触发了 Reddit 的速率限制，等待后重试
3. 某些 subreddit 可能需要认证

### 测试失败：429 Too Many Requests
```
Error: reddit returned status 429
```
**解决方案：** 
1. 增加请求之间的延迟（当前为 1 秒）
2. 减少并发请求数量
3. 等待一段时间后重试

## 贡献

如果你发现 bug 或有改进建议，请：
1. 添加相应的测试用例
2. 确保所有测试通过
3. 提交 PR

## 参考资料

- [Reddit RSS Feeds](https://www.reddit.com/wiki/rss)
- [gofeed 文档](https://github.com/mmcdole/gofeed)
- [Go Testing 最佳实践](https://go.dev/doc/tutorial/add-a-test)
