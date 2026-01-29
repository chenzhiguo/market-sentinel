# Market Sentinel Makefile
# -------------------------

# 变量定义
APP_NAME := sentinel
BIN_DIR := bin
CMD_DIR := ./cmd/sentinel
GO := go
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
.DEFAULT_GOAL := build

# 检查并创建bin目录
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

## build: 编译二进制文件到bin目录
.PHONY: build
build: $(BIN_DIR)
	@echo "==> Building $(APP_NAME)..."
	CGO_ENABLED=1 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "==> Binary built at $(BIN_DIR)/$(APP_NAME)"

## build-static: 静态编译 (适用于Alpine/Docker)
.PHONY: build-static
build-static: $(BIN_DIR)
	@echo "==> Building static binary..."
	CGO_ENABLED=1 $(GO) build -a -ldflags '-linkmode external -extldflags "-static" -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)' -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "==> Static binary built at $(BIN_DIR)/$(APP_NAME)"

## build-linux: 交叉编译Linux版本
.PHONY: build-linux
build-linux: $(BIN_DIR)
	@echo "==> Cross-compiling for Linux..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)
	@echo "==> Linux binary built at $(BIN_DIR)/$(APP_NAME)-linux-amd64"

## clean: 清理构建产物
.PHONY: clean
clean:
	@echo "==> Cleaning..."
	@rm -rf $(BIN_DIR)
	@echo "==> Cleaned"

## run: 运行应用
.PHONY: run
run: build
	@echo "==> Running $(APP_NAME)..."
	./$(BIN_DIR)/$(APP_NAME) serve

## test: 运行测试
.PHONY: test
test:
	@echo "==> Running tests..."
	$(GO) test -v ./...

## test-coverage: 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "==> Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage report generated at coverage.html"

## mod: 更新依赖
.PHONY: mod
mod:
	@echo "==> Updating dependencies..."
	$(GO) mod tidy
	$(GO) mod download
	@echo "==> Dependencies updated"

## fmt: 格式化代码
.PHONY: fmt
fmt:
	@echo "==> Formatting code..."
	$(GO) fmt ./...

## lint: 代码检查 (需要安装golangci-lint)
.PHONY: lint
lint:
	@echo "==> Running linter..."
	golangci-lint run ./...

## docker: 构建Docker镜像
.PHONY: docker
docker:
	@echo "==> Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "==> Docker image built: $(APP_NAME):$(VERSION)"

## docker-compose: 使用docker-compose启动
.PHONY: docker-compose
docker-compose:
	@echo "==> Starting with docker-compose..."
	docker-compose up -d

## help: 显示帮助信息
.PHONY: help
help:
	@echo "Market Sentinel - Available targets:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

.PHONY: all
all: clean build test
