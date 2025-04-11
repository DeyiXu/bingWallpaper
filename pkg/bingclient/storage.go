package bingclient

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage 是存储接口，用于保存数据到不同类型的存储介质
type Storage interface {
	// Save 保存数据到指定路径
	Save(data []byte, path string) error
	// SaveReader 从读取器保存数据到指定路径
	SaveReader(reader io.Reader, path string) error
	// Exists 检查路径是否存在
	Exists(path string) bool
}

// FileStorage 是一个文件系统存储实现
type FileStorage struct {
	Logger          Logger      // 日志记录器
	MkdirPermission os.FileMode // 目录创建权限
	FilePermission  os.FileMode // 文件创建权限
}

// NewFileStorage 创建一个新的文件存储实例
func NewFileStorage(logger Logger) *FileStorage {
	if logger == nil {
		logger = &NullLogger{}
	}

	return &FileStorage{
		Logger:          logger,
		MkdirPermission: 0755,
		FilePermission:  0644,
	}
}

// Save 将数据保存到指定路径的文件
func (fs *FileStorage) Save(data []byte, path string) error {
	fs.Logger.Debug("保存 %d 字节数据到文件: %s", len(data), path)

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, fs.MkdirPermission); err != nil {
		fs.Logger.Error("创建目录失败: %v", err)
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, fs.FilePermission); err != nil {
		fs.Logger.Error("写入文件失败: %v", err)
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fs.Logger.Info("成功保存数据到: %s", path)
	return nil
}

// SaveReader 从读取器保存数据到指定路径的文件
func (fs *FileStorage) SaveReader(reader io.Reader, path string) error {
	fs.Logger.Debug("从读取器保存数据到文件: %s", path)

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, fs.MkdirPermission); err != nil {
		fs.Logger.Error("创建目录失败: %v", err)
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FilePermission)
	if err != nil {
		fs.Logger.Error("创建文件失败: %v", err)
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入数据
	written, err := io.Copy(file, reader)
	if err != nil {
		fs.Logger.Error("写入文件失败: %v", err)
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fs.Logger.Info("成功保存 %d 字节数据到: %s", written, path)
	return nil
}

// Exists 检查路径是否存在
func (fs *FileStorage) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ImageFilenameGenerator 是生成图片文件名的接口
type ImageFilenameGenerator interface {
	// GenerateImageFilename 基于图片数据生成文件名
	GenerateImageFilename(imageData *ImageData, basePath string) string
	// GenerateJsonFilename 基于图片数据生成 JSON 文件名
	GenerateJsonFilename(imageData *ImageData, basePath string) string
}

// DefaultFilenameGenerator 是默认的文件名生成器
type DefaultFilenameGenerator struct {
	Logger Logger // 日志记录器
}

// NewDefaultFilenameGenerator 创建一个新的默认文件名生成器
func NewDefaultFilenameGenerator(logger Logger) *DefaultFilenameGenerator {
	if logger == nil {
		logger = &NullLogger{}
	}

	return &DefaultFilenameGenerator{
		Logger: logger,
	}
}

// GenerateImageFilename 根据图片数据生成图片文件名
func (g *DefaultFilenameGenerator) GenerateImageFilename(imageData *ImageData, basePath string) string {
	// 优先使用标题作为文件名
	description := imageData.Title

	// 如果标题为空，从版权信息中提取
	if description == "" {
		description = strings.Split(imageData.Copyright, "，")[0]
		description = strings.Split(description, "(")[0]
	}

	description = strings.TrimSpace(description)

	// 替换特殊字符
	description = strings.ReplaceAll(description, " ", "_")
	description = strings.ReplaceAll(description, "/", "-")
	description = strings.ReplaceAll(description, "\\", "-")
	description = strings.ReplaceAll(description, ":", "-")
	description = strings.ReplaceAll(description, "?", "")
	description = strings.ReplaceAll(description, "*", "")
	description = strings.ReplaceAll(description, "<", "")
	description = strings.ReplaceAll(description, ">", "")
	description = strings.ReplaceAll(description, "|", "-")
	description = strings.ReplaceAll(description, "\"", "'")

	// 生成文件名
	filename := fmt.Sprintf("%s_%s.jpg", imageData.Startdate, description)

	g.Logger.Debug("生成图片文件名: %s", filename)
	return filepath.Join(basePath, filename)
}

// GenerateJsonFilename 根据图片数据生成 JSON 文件名
func (g *DefaultFilenameGenerator) GenerateJsonFilename(imageData *ImageData, basePath string) string {
	// 使用日期作为文件名
	filename := fmt.Sprintf("bing_data_%s.json", imageData.Startdate)

	g.Logger.Debug("生成 JSON 文件名: %s", filename)
	return filepath.Join(basePath, filename)
}

// BingImageStorage 是 Bing 壁纸专用的存储工具
type BingImageStorage struct {
	Storage   Storage                // 存储实现
	Generator ImageFilenameGenerator // 文件名生成器
	OutputDir string                 // 输出目录
	Logger    Logger                 // 日志记录器
}

// NewBingImageStorage 创建一个新的 Bing 壁纸存储工具
func NewBingImageStorage(outputDir string, logger Logger) *BingImageStorage {
	if logger == nil {
		logger = &NullLogger{}
	}

	return &BingImageStorage{
		Storage:   NewFileStorage(logger),
		Generator: NewDefaultFilenameGenerator(logger),
		OutputDir: outputDir,
		Logger:    logger,
	}
}

// SaveImage 保存图片数据到文件
func (bis *BingImageStorage) SaveImage(data []byte, imageData *ImageData) (string, error) {
	bis.Logger.Info("保存图片数据...")

	// 生成文件路径
	filePath := bis.Generator.GenerateImageFilename(imageData, bis.OutputDir)

	// 保存数据
	err := bis.Storage.Save(data, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// SaveImageFromReader 从读取器保存图片数据
func (bis *BingImageStorage) SaveImageFromReader(reader io.Reader, imageData *ImageData) (string, error) {
	bis.Logger.Info("从读取器保存图片数据...")

	// 生成文件路径
	filePath := bis.Generator.GenerateImageFilename(imageData, bis.OutputDir)

	// 保存数据
	err := bis.Storage.SaveReader(reader, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// SaveJson 保存 JSON 数据到文件
func (bis *BingImageStorage) SaveJson(data []byte, imageData *ImageData) (string, error) {
	bis.Logger.Info("保存 JSON 数据...")

	// 生成文件路径
	filePath := bis.Generator.GenerateJsonFilename(imageData, bis.OutputDir)

	// 保存数据
	err := bis.Storage.Save(data, filePath)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// SetOutputDir 设置输出目录
func (bis *BingImageStorage) SetOutputDir(outputDir string) {
	bis.OutputDir = outputDir
}

// SetStorage 设置存储实现
func (bis *BingImageStorage) SetStorage(storage Storage) {
	bis.Storage = storage
}

// SetFilenameGenerator 设置文件名生成器
func (bis *BingImageStorage) SetFilenameGenerator(generator ImageFilenameGenerator) {
	bis.Generator = generator
}
