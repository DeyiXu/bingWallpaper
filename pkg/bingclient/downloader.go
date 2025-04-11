package bingclient

import (
	"fmt"
	"time"
)

// Downloader 是 Bing 壁纸下载器，协调 Client 与 Storage
type Downloader struct {
	Client       *Client           // API 客户端
	Storage      *BingImageStorage // 存储工具
	Logger       Logger            // 日志记录器
	SaveJsonData bool              // 是否保存JSON数据
}

// NewDownloader 创建新的壁纸下载器
func NewDownloader(client *Client, storage *BingImageStorage) *Downloader {
	// 使用 Client 的 GetLogger 方法获取日志记录器
	return &Downloader{
		Client:       client,
		Storage:      storage,
		Logger:       client.GetLogger(), // 通过方法获取 logger
		SaveJsonData: true,               // 默认保存 JSON 数据
	}
}

// SetLogger 设置自定义日志记录器
func (d *Downloader) SetLogger(logger Logger) {
	d.Logger = logger
}

// DownloadResult 壁纸下载结果
type DownloadResult struct {
	ImageData   ImageData // 图片元数据
	ImagePath   string    // 图片保存路径
	JsonPath    string    // JSON数据保存路径
	DownloadErr error     // 下载错误
	JsonErr     error     // JSON保存错误
}

// FetchAndSaveWallpaper 获取并保存单张壁纸
// daysAgo 指定获取多少天前的壁纸
func (d *Downloader) FetchAndSaveWallpaper(daysAgo int) (*DownloadResult, error) {

	d.Logger.Info("===== 开始处理 %d 天前的壁纸 =====", daysAgo)

	// 1. 获取图片元数据
	imageData, err := d.Client.FetchImageData(daysAgo)
	if err != nil {
		d.Logger.Error("获取图片数据失败: %v", err)
		return nil, fmt.Errorf("获取图片数据失败: %v", err)
	}

	// 使用另一个方法处理图片数据
	return d.SaveWallpaper(imageData, daysAgo)
}

// SaveWallpaper 保存单张壁纸
// 当已有 ImageData 时，可直接调用此方法
func (d *Downloader) SaveWallpaper(imageData *ImageData, daysAgo int) (*DownloadResult, error) {
	result := &DownloadResult{}
	result.ImageData = *imageData

	// 1. 下载并保存图片
	d.Logger.Info("下载并保存图片...")
	imageBytes, err := d.Client.FetchRawImageData(imageData)
	if err != nil {
		result.DownloadErr = err
		d.Logger.Warning("图片下载失败: %v", err)
		// 返回错误但同时也返回结果，以便调用者可以看到部分完成的结果
		return result, fmt.Errorf("图片下载失败: %v", err)
	}

	imagePath, err := d.Storage.SaveImage(imageBytes, imageData)
	if err != nil {
		result.DownloadErr = err
		d.Logger.Warning("图片保存失败: %v", err)
		return result, fmt.Errorf("图片保存失败: %v", err)
	}

	result.ImagePath = imagePath
	d.Logger.Info("图片已保存到: %s", imagePath)

	// 2. 只有在启用 SaveJsonData 时才获取并保存 JSON 数据
	if d.SaveJsonData {
		d.Logger.Info("下载并保存 JSON 数据...")
		jsonBytes, err := d.Client.FetchRawJsonData(d.Client.GetBingApiURL(daysAgo, 1))
		if err != nil {
			result.JsonErr = err
			d.Logger.Warning("JSON 数据获取失败: %v", err)
			// 图片已成功保存，即使 JSON 失败也算基本成功，所以这里不返回错误
		} else {
			jsonPath, err := d.Storage.SaveJson(jsonBytes, imageData)
			if err != nil {
				result.JsonErr = err
				d.Logger.Warning("JSON 数据保存失败: %v", err)
			} else {
				result.JsonPath = jsonPath
				d.Logger.Info("JSON 数据已保存到: %s", jsonPath)
			}
		}
	} else {
		d.Logger.Debug("跳过 JSON 数据保存（已禁用）")
	}

	d.Logger.Info("===== 壁纸处理完成 =====")
	return result, nil
}

// FetchAndSaveWallpapers 获取并保存多天的壁纸
// continueOnError 控制遇到错误时是否继续处理其他壁纸
func (d *Downloader) FetchAndSaveWallpapers(days int, continueOnError bool) ([]*DownloadResult, error) {
	results := make([]*DownloadResult, 0, days)
	var lastError error

	d.Logger.Info("开始处理最近 %d 天的壁纸", days)

	for i := 0; i < days; i++ {
		result, err := d.FetchAndSaveWallpaper(i)
		if err != nil {
			d.Logger.Error("处理第 %d 天的壁纸失败: %v", i, err)
			lastError = fmt.Errorf("处理第 %d 天的壁纸失败: %v", i, err)

			if !continueOnError {
				return results, lastError
			}
			// 如果需要继续，将结果添加到列表中，即使有错误
			if result != nil {
				results = append(results, result)
			}
		} else {
			results = append(results, result)
		}

		// 避免请求过于频繁
		if i < days-1 {
			d.Logger.Debug("等待1秒后继续...")
			time.Sleep(1 * time.Second)
		}
	}

	d.Logger.Info("所有壁纸处理完成！共 %d 张，成功 %d 张", days, len(results))

	// 如果启用了继续处理且有错误，返回最后一个错误
	if lastError != nil && continueOnError {
		return results, fmt.Errorf("有部分壁纸处理失败: %v", lastError)
	}

	return results, nil
}

// SaveWallpapers 保存多张壁纸
// 当已有 ImageData 列表时，可直接调用此方法
// continueOnError 控制遇到错误时是否继续处理其他壁纸
func (d *Downloader) SaveWallpapers(imageDataList []ImageData, continueOnError bool) ([]*DownloadResult, error) {
	results := make([]*DownloadResult, 0, len(imageDataList))
	var lastError error

	d.Logger.Info("开始处理 %d 张壁纸", len(imageDataList))

	for i, imageData := range imageDataList {
		// 为了找到正确的 daysAgo 值，我们假设列表是按照时间顺序排列的
		daysAgo := i

		result, err := d.SaveWallpaper(&imageData, daysAgo)
		if err != nil {
			d.Logger.Error("处理第 %d 张壁纸失败: %v", i, err)
			lastError = fmt.Errorf("处理第 %d 张壁纸失败: %v", i, err)

			if !continueOnError {
				return results, lastError
			}
			// 如果需要继续，将结果添加到列表中，即使有错误
			if result != nil {
				results = append(results, result)
			}
		} else {
			results = append(results, result)
		}

		// 避免请求过于频繁
		if i < len(imageDataList)-1 {
			d.Logger.Debug("等待1秒后继续...")
			time.Sleep(1 * time.Second)
		}
	}

	d.Logger.Info("所有壁纸处理完成！共处理 %d 张，成功 %d 张", len(imageDataList), len(results))

	// 如果启用了继续处理且有错误，返回最后一个错误
	if lastError != nil && continueOnError {
		return results, fmt.Errorf("有部分壁纸处理失败: %v", lastError)
	}

	return results, nil
}

// DownloadLatestWallpapers 批量下载最新壁纸的优化方法
// 这个方法会一次获取多天的数据，然后批量处理，减少 API 请求次数
// continueOnError 控制遇到错误时是否继续处理其他壁纸
func (d *Downloader) DownloadLatestWallpapers(days int, continueOnError bool) ([]*DownloadResult, error) {
	if days <= 0 || days > 16 {
		return nil, fmt.Errorf("days 必须在 1-16 之间，当前值: %d", days)
	}

	d.Logger.Info("正在批量获取最近 %d 天的壁纸", days)

	// 1. 一次性获取多天的壁纸数据
	imagesData, err := d.Client.FetchMultipleImageData(days)
	if err != nil {
		d.Logger.Error("获取壁纸数据失败: %v", err)
		return nil, err
	}

	// 2. 批量保存壁纸，使用传入的 continueOnError 参数
	return d.SaveWallpapers(imagesData, continueOnError)
}
