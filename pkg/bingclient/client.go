package bingclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// API响应结构体
type HPImageArchiveResponse struct {
	Images   []ImageData `json:"images"`
	Tooltips Tooltips    `json:"tooltips"`
}

// 图片数据结构体
type ImageData struct {
	Startdate     string   `json:"startdate"`     // 图片日期
	Fullstartdate string   `json:"fullstartdate"` // 完整日期时间
	Enddate       string   `json:"enddate"`       // 结束日期
	URL           string   `json:"url"`           // 图片相对路径
	Urlbase       string   `json:"urlbase"`       // 基础URL
	Copyright     string   `json:"copyright"`     // 版权信息
	Copyrightlink string   `json:"copyrightlink"` // 版权链接
	Title         string   `json:"title"`         // 标题
	Quiz          string   `json:"quiz"`          // 测验链接
	Wp            bool     `json:"wp"`            // 是否为壁纸
	Hsh           string   `json:"hsh"`           // 哈希值
	Drk           int      `json:"drk"`           // 暗色值
	Top           int      `json:"top"`           // 顶部位置
	Bot           int      `json:"bot"`           // 底部位置
	Hs            []string `json:"hs"`            // 热点区域
}

// 提示信息结构体
type Tooltips struct {
	Loading  string `json:"loading"`  // 加载提示
	Previous string `json:"previous"` // 上一个提示
	Next     string `json:"next"`     // 下一个提示
	Walle    string `json:"walle"`    // 壁纸不可用提示
	Walls    string `json:"walls"`    // 壁纸下载提示
}

// 客户端配置选项
type ClientOption func(*Client)

// Bing壁纸客户端
type Client struct {
	baseURL     string        // API基础URL
	timeout     time.Duration // 超时时间
	userAgent   string        // 用户代理
	locale      string        // 语言区域
	highQuality bool          // 高清质量
	logger      Logger        // 日志记录器
	httpClient  *http.Client  // HTTP客户端
}

// 创建新的客户端实例
func NewClient(options ...ClientOption) *Client {
	// 默认设置
	client := &Client{
		baseURL:     "https://www.bing.com/HPImageArchive.aspx",
		timeout:     10 * time.Second,
		userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		locale:      "zh-CN",
		highQuality: true,
		logger:      NewLogger(), // 使用默认日志记录器
	}

	// 应用配置选项
	for _, option := range options {
		option(client)
	}

	// 初始化HTTP客户端
	client.httpClient = &http.Client{
		Timeout: client.timeout,
	}

	return client
}

// 设置日志记录器选项
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// 设置超时时间选项
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// 设置是否使用高质量图片选项
func WithHighQuality(enabled bool) ClientOption {
	return func(c *Client) {
		c.highQuality = enabled
	}
}

// 设置语言区域选项
func WithLocale(locale string) ClientOption {
	return func(c *Client) {
		c.locale = locale
	}
}

// 设置自定义用户代理选项
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// 设置自定义基础URL选项
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// 发送HTTP请求并返回响应体
func (c *Client) sendRequest(method, url string) ([]byte, error) {
	c.logger.Debug("发送 %s 请求到 %s", method, url)

	// 创建请求
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		c.logger.Error("创建请求失败: %v", err)
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", c.userAgent)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("HTTP错误状态码: %d", resp.StatusCode)
		return nil, fmt.Errorf("HTTP错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("读取响应失败: %v", err)
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	c.logger.Debug("成功收到响应 (%d 字节)", len(body))
	return body, nil
}

// GetBingImageURL 获取 Bing 图片的完整 URL
func (c *Client) GetBingImageURL(imageData *ImageData) string {
	// 构建完整图片URL
	imageURL := fmt.Sprintf("https://www.bing.com%s", imageData.URL)

	// 如果需要高质量图片，替换为UHD版本
	if c.highQuality {
		imageURL = strings.Replace(imageURL, "1920x1080", "UHD", 1)
	}

	return imageURL
}

// GetBingApiURL 获取 Bing API 的完整 URL
func (c *Client) GetBingApiURL(daysAgo, count int) string {
	return fmt.Sprintf("%s?format=js&n=%d&idx=%d&mkt=%s", c.baseURL, count, daysAgo, c.locale)
}

// GetLogger 返回客户端使用的日志记录器
func (c *Client) GetLogger() Logger {
	return c.logger
}

// parseImageResponse 解析 API 响应数据
func (c *Client) parseImageResponse(data []byte) ([]ImageData, error) {
	c.logger.Debug("正在解析 API 响应数据...")

	var archiveResp HPImageArchiveResponse
	if err := json.Unmarshal(data, &archiveResp); err != nil {
		c.logger.Error("JSON解析失败: %v", err)
		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	if len(archiveResp.Images) == 0 {
		c.logger.Error("未找到图片数据")
		return nil, fmt.Errorf("未找到图片数据")
	}

	c.logger.Debug("成功解析 %d 条图片数据", len(archiveResp.Images))
	return archiveResp.Images, nil
}

// FetchImageData 获取指定日期的壁纸数据
func (c *Client) FetchImageData(daysAgo int) (*ImageData, error) {
	// 使用通用解析方法解析响应
	images, err := c.fetchMultipleImageData(daysAgo, 1)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		c.logger.Error("未找到图片数据")
		return nil, fmt.Errorf("未找到图片数据")
	}
	c.logger.Info("成功获取壁纸数据")
	c.logger.Debug("壁纸标题: %s", images[0].Title)
	return &images[0], nil
}

// FetchRawImageData 获取原始图片数据
func (c *Client) FetchRawImageData(imageData *ImageData) ([]byte, error) {
	imageURL := c.GetBingImageURL(imageData)
	c.logger.Info("获取图片数据: %s", imageURL)

	return c.sendRequest("GET", imageURL)
}

// FetchRawJsonData 获取原始的 JSON 数据
func (c *Client) FetchRawJsonData(apiURL string) ([]byte, error) {
	c.logger.Info("获取 JSON 数据: %s", apiURL)

	return c.sendRequest("GET", apiURL)
}

// FetchMultipleImageData 获取多天的壁纸数据
func (c *Client) FetchMultipleImageData(days int) ([]ImageData, error) {
	if days <= 0 || days > 16 {
		return nil, fmt.Errorf("days 必须在 1-16 之间，当前值: %d", days)
	}
	return c.fetchMultipleImageData(0, days)
}

// fetchMultipleImageData 获取多天的壁纸数据
// 内部方法，供 FetchImageData 和 FetchMultipleImageData 使用
func (c *Client) fetchMultipleImageData(daysAgo int, count int) ([]ImageData, error) {
	apiURL := c.GetBingApiURL(daysAgo, count)
	c.logger.Info("正在获取壁纸数据: daysAgo=%d, count=%d, URL=%s", daysAgo, count, apiURL)

	body, err := c.FetchRawJsonData(apiURL)
	if err != nil {
		return nil, err
	}

	// 使用通用解析方法解析响应
	images, err := c.parseImageResponse(body)
	if err != nil {
		return nil, err
	}

	c.logger.Info("成功获取 %d 天的壁纸数据", len(images))
	return images, nil
}
