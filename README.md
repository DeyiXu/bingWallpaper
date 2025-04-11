# Bing 壁纸下载工具

一个用 Go 语言编写的简单、高效的必应（Bing）每日壁纸下载工具。该工具可以获取 Bing 首页的精美壁纸及其元数据，并支持多种配置选项。

## 功能特点

- 下载 Bing 首页每日壁纸（支持最近 16 天的壁纸）
- 可选下载高清版本（UHD）或标准版本
- 支持保存图片元数据（JSON 格式）
- 自动生成基于日期和图片描述的文件名
- 支持多语言区域设置
- 命令行界面，简单易用
- 完善的日志记录系统，支持多种日志级别
- 可作为客户端库在其他 Go 项目中集成使用

## 项目结构

```
./
├── main.go                 # 主程序
├── go.mod                  # Go模块定义
├── Makefile                # 编译构建配置
├── README.md               # 项目说明文档
├── bin/                    # 编译输出目录
│   ├── bingWallpaper_linux_amd64    # Linux amd64 可执行文件
│   ├── bingWallpaper_windows_amd64.exe  # Windows amd64 可执行文件
│   └── ...                 # 其他平台的可执行文件
├── bing_wallpapers/        # 壁纸保存默认目录
│   ├── YYYYMMDD_描述.jpg    # 下载的壁纸
│   └── bing_data_YYYYMMDD.json  # 元数据
└── pkg/
    └── bingclient/         # 客户端包
        ├── client.go       # 客户端核心功能
        ├── downloader.go   # 下载器实现
        ├── logger.go       # 日志接口系统
        ├── storage.go      # 存储实现
        └── utils.go        # 工具函数
```

## 安装

### 前提条件

- Go 1.16 或更高版本

### 使用 Makefile 编译

```bash
# 编译当前平台版本
make

# 交叉编译所有支持的平台版本
make cross-build

# 编译特定平台版本
make linux-amd64
make windows-amd64
make darwin-arm64

# 清理编译输出
make clean

# 查看帮助
make help
```

### 从源码安装

```bash
git clone https://github.com/DeyiXu/bingWallpaper.git
cd bingWallpaper
go build
```

## 使用方法

### 命令行使用

基本用法：

```bash
# 使用默认配置下载最近7天的壁纸
./bingWallpaper
```

带参数使用：

```bash
# 下载最近10天的壁纸到指定目录，使用高清版本并保存JSON数据
./bingWallpaper -days 10 -dir ./my_wallpapers -hd=true -json=true

# 下载英文区域的壁纸
./bingWallpaper -locale en-US

# 只下载最新的1张壁纸
./bingWallpaper -days 1

# 查看调试级别的详细日志
./bingWallpaper -log-level=debug

# 不显示日志中的时间戳
./bingWallpaper -no-time

# 显示版本信息并退出
./bingWallpaper -version
```

### 参数说明

| 参数 | 默认值 | 描述 |
|------|--------|------|
| `-dir` | `./bing_wallpapers` | 壁纸保存目录 |
| `-days` | `7` | 下载最近几天的壁纸 (1-16) |
| `-hd` | `true` | 是否下载高清壁纸 |
| `-json` | `false` | 是否保存原始JSON数据 |
| `-locale` | `zh-CN` | 语言区域 (如 zh-CN, en-US, ja-JP 等) |
| `-log-level` | `info` | 日志级别 (debug, info, warning, error) |
| `-no-time` | `false` | 日志中不显示时间戳 |
| `-version` | `false` | 显示版本信息并退出 |

### 版本信息

程序启动时会显示版本、构建时间和Git提交哈希信息，便于跟踪和调试：

```
BingWallpaper 版本: 1.0.0 (构建于: 2025-04-11 09:02:51, 提交: 1b45578)
```

也可以使用 `-version` 参数仅显示版本信息而不执行下载功能。

## 作为库使用

您可以在自己的 Go 项目中引入此库：

```go
package main

import (
	"fmt"
	"time"

	"github.com/DeyiXu/bingWallpaper/pkg/bingclient"
)

func main() {
	// 创建客户端，配置选项
	client := bingclient.NewClient(
		bingclient.WithHighQuality(true),
		bingclient.WithLocale("zh-CN"),
		bingclient.WithTimeout(15*time.Second),
	)

	// 下载最新的3张壁纸
	results, err := client.DownloadWallpapers(3)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}

	// 处理下载结果
	for i, result := range results {
		fmt.Printf("第 %d 张壁纸:\n", i+1)
		fmt.Println(bingclient.GetImageSummary(&result.ImageData))
		fmt.Printf("保存路径: %s\n\n", result.ImagePath)
	}
}
```

### 客户端选项

在创建客户端时，可以配置以下选项：

```go
// 所有可用的配置选项
client := bingclient.NewClient(
	// 是否下载高清版本
	bingclient.WithHighQuality(true),
	
	// 设置语言区域
	bingclient.WithLocale("zh-CN"),
	
	// 设置请求超时时间
	bingclient.WithTimeout(15*time.Second),
	
	// 设置自定义用户代理
	bingclient.WithUserAgent("My Bing Wallpaper App/1.0"),
	
	// 设置自定义API基础URL（如果必要）
	bingclient.WithBaseURL("https://www.bing.com/HPImageArchive.aspx"),
	
	// 设置自定义日志记录器
	bingclient.WithLogger(customLogger),
)
```

### API参考

#### 核心方法

- `NewClient(options ...ClientOption) *Client` - 创建新的客户端实例
- `client.FetchImageData(daysAgo int) (*ImageData, error)` - 获取指定日期的壁纸数据
- `client.DownloadWallpaper(daysAgo int) (*DownloadResult, error)` - 下载指定日期的壁纸
- `client.DownloadWallpapers(days int) ([]*DownloadResult, error)` - 下载多天的壁纸
- `client.GetLogger() Logger` - 获取客户端的日志记录器

#### 工具函数

- `GetImageSummary(imageData *ImageData) string` - 获取图片信息的简要描述
- `FormatDate(dateStr string) (string, error)` - 格式化日期字符串为可读形式
- `IsImageFromToday(imageData *ImageData) bool` - 检查图片是否是今天的

## 日志系统

项目实现了灵活的日志接口系统，支持不同级别的日志记录：

- **Debug**: 调试信息，包含详细的处理过程
- **Info**: 一般信息，记录主要操作
- **Warning**: 警告信息，非致命但需注意的问题
- **Error**: 错误信息，影响程序正常运行的问题

### 自定义日志记录器

您可以创建自定义日志记录器来控制日志行为：

```go
// 创建自定义日志记录器
logger := bingclient.NewLogger(
    // 设置日志级别
    bingclient.WithLevel(bingclient.LogLevelDebug),
    
    // 设置输出目标（默认为标准输出）
    bingclient.WithWriter(os.Stderr),
    
    // 是否显示时间戳
    bingclient.WithTimeDisplay(true),
    
    // 是否显示日志级别标签
    bingclient.WithLevelDisplay(true),
)

// 创建客户端时使用自定义日志记录器
client := bingclient.NewClient(
    bingclient.WithLogger(logger),
    // 其他选项...
)
```

### 实现自定义 Logger 接口

您还可以完全自定义日志行为，只需实现 `Logger` 接口：

```go
type MyCustomLogger struct {
    // 您的自定义字段
}

// 实现 Logger 接口的方法
func (l *MyCustomLogger) Debug(format string, args ...interface{}) {
    // 自定义调试级别日志处理
}

func (l *MyCustomLogger) Info(format string, args ...interface{}) {
    // 自定义信息级别日志处理
}

func (l *MyCustomLogger) Warning(format string, args ...interface{}) {
    // 自定义警告级别日志处理
}

func (l *MyCustomLogger) Error(format string, args ...interface{}) {
    // 自定义错误级别日志处理
}

func (l *MyCustomLogger) SetLevel(level bingclient.LogLevel) {
    // 设置日志级别
}

func (l *MyCustomLogger) GetLevel() bingclient.LogLevel {
    // 获取当前日志级别
    return bingclient.LogLevelInfo
}

// 使用自定义日志记录器
client := bingclient.NewClient(
    bingclient.WithLogger(&MyCustomLogger{}),
)
```

## 示例与用例

### 1. 每天自动下载最新壁纸

```go
package main

import (
	"fmt"
	"github.com/DeyiXu/bingWallpaper/pkg/bingclient"
)

func main() {
	client := bingclient.NewClient()
	
	// 只下载今天的壁纸（daysAgo=0 表示今天）
	result, err := client.DownloadWallpaper(0)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}
	
	fmt.Printf("今日壁纸已下载: %s\n", result.ImagePath)
}
```

### 2. 获取壁纸元数据不下载图片

```go
package main

import (
	"fmt"
	"github.com/DeyiXu/bingWallpaper/pkg/bingclient"
)

func main() {
	client := bingclient.NewClient()
	
	// 获取最近7天的壁纸数据
	for i := 0; i < 7; i++ {
		imageData, err := client.FetchImageData(i)
		if err != nil {
			fmt.Printf("获取第 %d 天数据失败: %v\n", i, err)
			continue
		}
		
		// 打印壁纸信息
		fmt.Printf("第 %d 天壁纸:\n", i)
		fmt.Printf("  标题: %s\n", imageData.Title)
		fmt.Printf("  描述: %s\n", imageData.Copyright)
		fmt.Printf("  日期: %s\n\n", imageData.Startdate)
	}
}
```

### 3. 自定义文件命名

```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"github.com/DeyiXu/bingWallpaper/pkg/bingclient"
)

func main() {
	client := bingclient.NewClient()
	
	// 获取今天的壁纸数据
	imageData, err := client.FetchImageData(0)
	if err != nil {
		fmt.Printf("获取数据失败: %v\n", err)
		return
	}
	
	// 构建图片URL
	imageURL := fmt.Sprintf("https://www.bing.com%s", imageData.URL)
	
	// 自定义文件名
	fileName := fmt.Sprintf("bing_%s_%s.jpg", imageData.Startdate, imageData.Hsh)
	savePath := filepath.Join("./custom_wallpapers", fileName)
	
	// 确保目录存在
	os.MkdirAll(filepath.Dir(savePath), 0755)
	
	// 下载图片
	resp, err := http.Get(imageURL)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 保存图片
	file, err := os.Create(savePath)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("保存文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("图片已保存至: %s\n", savePath)
}
```

### 4. 配置日志级别的示例

```go
package main

import (
	"fmt"
	"os"
	"github.com/DeyiXu/bingWallpaper/pkg/bingclient"
)

func main() {
	// 创建调试级别的日志记录器
	logger := bingclient.NewLogger(
		bingclient.WithLevel(bingclient.LogLevelDebug),
		bingclient.WithWriter(os.Stdout),
	)
	
	client := bingclient.NewClient(
		bingclient.WithLogger(logger),
	)
	
	// 下载最新壁纸（将显示详细的调试信息）
	result, err := client.DownloadWallpaper(0)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}
	
	fmt.Printf("壁纸已下载: %s\n", result.ImagePath)
}
```

## 常见问题解答

**Q: 为什么有时候会下载失败？**  
A: 这可能是因为网络连接问题或 Bing API 临时不可用。请稍后再试。

**Q: 如何设置定时任务自动下载每天的壁纸？**  
A: 您可以使用操作系统的任务调度工具，如 Linux 的 crontab 或 Windows 的任务计划程序，设置每天运行此工具。

**Q: 支持代理服务器吗？**  
A: 目前不直接支持代理配置，但您可以通过系统环境变量 `HTTP_PROXY` 和 `HTTPS_PROXY` 设置代理。

## 许可证

MIT 许可证

## 致谢

感谢 Microsoft Bing 提供精美的每日壁纸。