# bingWallpaper 编译构建 Makefile

# 变量定义
PACKAGE_NAME := bingWallpaper
OUTPUT_DIR := bin
GO := go
GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH ?= $(shell go env GOARCH)
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date "+%Y-%m-%d\ %H:%M:%S")
COMMIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Windows 特殊处理
ifeq ($(GOOS),windows)
	OUTPUT_FILE := $(OUTPUT_DIR)/$(PACKAGE_NAME)_$(GOOS)_$(GOARCH).exe
else
	OUTPUT_FILE := $(OUTPUT_DIR)/$(PACKAGE_NAME)_$(GOOS)_$(GOARCH)
endif

# LDFLAGS 设置 - 正确转义双引号和处理空格
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X 'main.BuildTime=$(BUILD_TIME)' -X main.CommitSHA=$(COMMIT_SHA)"

# 默认目标
.PHONY: all
all: build

# 创建输出目录
$(OUTPUT_DIR):
	mkdir -p $@

# 构建目标
.PHONY: build
build: $(OUTPUT_DIR)
	@echo "正在编译 $(PACKAGE_NAME) 版本 $(VERSION)"
	@echo "目标平台: $(GOOS)_$(GOARCH)"
	@echo "输出文件: $(OUTPUT_FILE)"
	$(GO) build $(LDFLAGS) -o $(OUTPUT_FILE) main.go
	@echo "编译成功: $(OUTPUT_FILE)"
	@chmod +x $(OUTPUT_FILE)

# 清理目标
.PHONY: clean
clean:
	@echo "清理构建目录..."
	rm -rf $(OUTPUT_DIR)
	@echo "清理完成"

# 交叉编译目标
.PHONY: cross-build
cross-build: linux-amd64 linux-arm64 windows-amd64 darwin-amd64 darwin-arm64

.PHONY: linux-amd64
linux-amd64:
	@echo "构建 Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: linux-arm64
linux-arm64:
	@echo "构建 Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(MAKE) build

.PHONY: windows-amd64
windows-amd64:
	@echo "构建 Windows (amd64)..."
	GOOS=windows GOARCH=amd64 $(MAKE) build

.PHONY: darwin-amd64
darwin-amd64:
	@echo "构建 macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(MAKE) build

.PHONY: darwin-arm64
darwin-arm64:
	@echo "构建 macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(MAKE) build

# 帮助信息
.PHONY: help
help:
	@echo "bingWallpaper 构建系统"
	@echo ""
	@echo "可用命令:"
	@echo "  make          - 构建当前平台的可执行文件"
	@echo "  make clean    - 清理构建目录"
	@echo "  make cross-build - 交叉编译所有支持的平台"
	@echo "  make linux-amd64 - 构建 Linux amd64 版本"
	@echo "  make linux-arm64 - 构建 Linux arm64 版本"
	@echo "  make windows-amd64 - 构建 Windows amd64 版本"
	@echo "  make darwin-amd64 - 构建 macOS amd64 版本"
	@echo "  make darwin-arm64 - 构建 macOS arm64 版本"
	@echo ""
	@echo "自定义环境变量:"
	@echo "  GOOS   - 目标操作系统 (linux, windows, darwin 等)"
	@echo "  GOARCH - 目标架构 (amd64, arm64 等)"