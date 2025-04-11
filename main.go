package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DeyiXu/bingWallpaper/pkg/bingclient"
)

// 由编译时 -ldflags 参数传入的值
var (
	Version   = "dev"
	BuildTime = "unknown"
	CommitSHA = "unknown"
)

func main() {
	// 命令行参数
	var (
		outputDir   string
		days        int
		highQuality bool
		saveJson    bool
		locale      string
		logLevel    string
		noTime      bool
		showVersion bool
	)

	flag.StringVar(&outputDir, "dir", "./bing_wallpapers", "壁纸保存目录")
	flag.IntVar(&days, "days", 7, "下载最近几天的壁纸 (1-16)")
	flag.BoolVar(&highQuality, "hd", true, "下载高清壁纸")
	flag.BoolVar(&saveJson, "json", false, "保存原始JSON数据")
	flag.StringVar(&locale, "locale", "zh-CN", "语言区域 (zh-CN, en-US, ja-JP 等)")
	flag.StringVar(&logLevel, "log-level", "info", "日志级别 (debug, info, warning, error)")
	flag.BoolVar(&noTime, "no-time", false, "日志中不显示时间戳")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息并退出")
	flag.Parse()

	// 处理版本信息显示请求
	if showVersion {
		// 显示版本信息
		fmt.Printf("BingWallpaper 版本: %s (构建于: %s, 提交: %s)\n\n", Version, BuildTime, CommitSHA)
		// 已经在开始显示了版本信息，所以这里直接退出
		os.Exit(0)
	}

	// 校验参数
	if days < 1 || days > 16 {
		fmt.Printf("错误: days参数必须在1到16之间\n")
		os.Exit(1)
	}

	// 获取绝对路径
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		fmt.Printf("错误: 无法获取绝对路径: %v\n", err)
		os.Exit(1)
	}

	// 设置日志级别
	var level bingclient.LogLevel
	switch logLevel {
	case "debug":
		level = bingclient.LogLevelDebug
	case "info":
		level = bingclient.LogLevelInfo
	case "warning":
		level = bingclient.LogLevelWarning
	case "error":
		level = bingclient.LogLevelError
	default:
		fmt.Printf("警告: 无效的日志级别 '%s'，使用默认级别 'info'\n", logLevel)
		level = bingclient.LogLevelInfo
	}

	// 创建日志记录器
	logger := bingclient.NewLogger(
		bingclient.WithLevel(level),
		bingclient.WithTimeDisplay(!noTime),
	)

	// 创建 Bing 壁纸客户端
	client := bingclient.NewClient(
		bingclient.WithHighQuality(highQuality),
		bingclient.WithLocale(locale),
		bingclient.WithTimeout(15*time.Second),
		bingclient.WithLogger(logger),
	)

	// 创建存储工具
	storage := bingclient.NewBingImageStorage(absOutputDir, logger)

	// 创建下载器
	downloader := bingclient.NewDownloader(client, storage)

	// 下载壁纸（使用优化的批量下载方法）
	results, err := downloader.DownloadLatestWallpapers(days, true)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 输出结果摘要
	var success, failed int
	for _, result := range results {
		if result.DownloadErr == nil {
			success++
		} else {
			failed++
		}
	}

	fmt.Printf("\n下载完成: 成功%d张，失败%d张\n", success, failed)
}
