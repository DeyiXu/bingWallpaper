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

// 自定义文件名生成器，用于支持指定文件名
type CustomFilenameGenerator struct {
	bingclient.DefaultFilenameGenerator
	CustomFilename string
}

// 重写生成图片文件名的方法
func (g *CustomFilenameGenerator) GenerateImageFilename(imageData *bingclient.ImageData, basePath string) string {
	// 如果指定了自定义文件名，则使用它
	if g.CustomFilename != "" {
		// 确保有正确的扩展名
		if filepath.Ext(g.CustomFilename) == "" {
			g.CustomFilename += ".jpg"
		}
		return filepath.Join(basePath, g.CustomFilename)
	}

	// 否则使用默认生成器的方法
	return g.DefaultFilenameGenerator.GenerateImageFilename(imageData, basePath)
}

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
		lastOnly    bool
		customName  string
		overwrite   bool
	)

	flag.StringVar(&outputDir, "dir", "./bing_wallpapers", "壁纸保存目录")
	flag.IntVar(&days, "days", 7, "下载最近几天的壁纸 (1-16)")
	flag.BoolVar(&highQuality, "hd", true, "下载高清壁纸")
	flag.BoolVar(&saveJson, "json", false, "保存原始JSON数据")
	flag.StringVar(&locale, "locale", "zh-CN", "语言区域 (zh-CN, en-US, ja-JP 等)")
	flag.StringVar(&logLevel, "log-level", "info", "日志级别 (debug, info, warning, error)")
	flag.BoolVar(&noTime, "no-time", false, "日志中不显示时间戳")
	flag.BoolVar(&showVersion, "version", false, "显示版本信息并退出")
	flag.BoolVar(&lastOnly, "last", false, "仅下载最后一天的壁纸")
	flag.StringVar(&customName, "name", "", "指定保存的文件名 (如 my-wallpaper.jpg)")
	flag.BoolVar(&overwrite, "overwrite", false, "如果文件已存在则覆盖")
	flag.Parse()

	// 处理版本信息显示请求
	if showVersion {
		// 显示版本信息
		fmt.Printf("BingWallpaper 版本: %s (构建于: %s, 提交: %s)\n\n", Version, BuildTime, CommitSHA)
		// 已经在开始显示了版本信息，所以这里直接退出
		os.Exit(0)
	}

	// 如果启用了仅下载最后一天，则强制设置 days 为 1
	if lastOnly {
		days = 1
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

	// 如果指定了自定义文件名，设置自定义文件名生成器
	if customName != "" {
		customGenerator := &CustomFilenameGenerator{
			DefaultFilenameGenerator: *bingclient.NewDefaultFilenameGenerator(logger),
			CustomFilename:           customName,
		}
		storage.SetFilenameGenerator(customGenerator)

		// 检查文件是否已存在且未指定覆盖
		if !overwrite {
			filePath := filepath.Join(absOutputDir, customName)
			if filepath.Ext(filePath) == "" {
				filePath += ".jpg"
			}

			if fileExists(filePath) {
				fmt.Printf("错误: 文件 %s 已存在。使用 -overwrite 选项覆盖现有文件。\n", filePath)
				os.Exit(1)
			}
		}
	}

	// 创建下载器
	downloader := bingclient.NewDownloader(client, storage)
	// 设置是否保存JSON数据
	downloader.SaveJsonData = saveJson

	var results []*bingclient.DownloadResult
	var downloadErr error

	// 根据是否只下载最后一天来选择下载方法
	if lastOnly {
		logger.Info("仅下载最后一天的壁纸")
		result, err := downloader.FetchAndSaveWallpaper(0)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			os.Exit(1)
		}
		results = []*bingclient.DownloadResult{result}
	} else {
		// 下载壁纸（使用优化的批量下载方法）
		results, downloadErr = downloader.DownloadLatestWallpapers(days, true)
		if downloadErr != nil {
			fmt.Printf("错误: %v\n", downloadErr)
			os.Exit(1)
		}
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

	// 如果只下载了一张，显示更详细的信息
	if lastOnly && len(results) > 0 && results[0].DownloadErr == nil {
		result := results[0]
		fmt.Printf("\n壁纸详情:\n")
		fmt.Printf("标题: %s\n", result.ImageData.Title)
		if formattedDate, err := bingclient.FormatDate(result.ImageData.Startdate); err == nil {
			fmt.Printf("日期: %s\n", formattedDate)
		}
		fmt.Printf("描述: %s\n", result.ImageData.Copyright)
		fmt.Printf("保存路径: %s\n", result.ImagePath)
		if saveJson && result.JsonPath != "" {
			fmt.Printf("元数据: %s\n", result.JsonPath)
		}
	}
}

// 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
