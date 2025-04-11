package bingclient

import (
	"fmt"
	"strings"
	"time"
)

// 格式化日期字符串为可读形式
func FormatDate(dateStr string) (string, error) {
	if len(dateStr) != 8 {
		return "", fmt.Errorf("无效的日期格式: %s", dateStr)
	}

	year := dateStr[0:4]
	month := dateStr[4:6]
	day := dateStr[6:8]

	return fmt.Sprintf("%s年%s月%s日", year, month, day), nil
}

// 格式化完整日期时间为可读形式
func FormatFullDateTime(fullDateStr string) (string, error) {
	if len(fullDateStr) < 12 {
		return "", fmt.Errorf("无效的完整日期时间格式: %s", fullDateStr)
	}

	year := fullDateStr[0:4]
	month := fullDateStr[4:6]
	day := fullDateStr[6:8]
	hour := fullDateStr[8:10]
	minute := fullDateStr[10:12]

	return fmt.Sprintf("%s年%s月%s日 %s:%s", year, month, day, hour, minute), nil
}

// 获取图片信息的简要描述
func GetImageSummary(imageData *ImageData) string {
	var builder strings.Builder

	// 添加标题
	if imageData.Title != "" {
		builder.WriteString(fmt.Sprintf("标题: %s\n", imageData.Title))
	}

	// 添加版权信息
	if imageData.Copyright != "" {
		builder.WriteString(fmt.Sprintf("描述: %s\n", imageData.Copyright))
	}

	// 添加日期
	if imageData.Startdate != "" {
		if date, err := FormatDate(imageData.Startdate); err == nil {
			builder.WriteString(fmt.Sprintf("日期: %s\n", date))
		}
	}

	return strings.TrimSpace(builder.String())
}

// 获取当前日期的字符串表示 (格式: YYYYMMDD)
func GetTodayDateString() string {
	now := time.Now()
	return now.Format("20060102")
}

// 检查图片是否是今天的
func IsImageFromToday(imageData *ImageData) bool {
	today := GetTodayDateString()
	return imageData.Startdate == today
}

// 提取壁纸描述（用于文件名）
func ExtractWallpaperDescription(imageData *ImageData) string {
	// 优先使用标题
	description := imageData.Title

	// 如果标题为空，使用版权信息
	if description == "" {
		description = strings.Split(imageData.Copyright, "，")[0]
		description = strings.Split(description, "(")[0]
	}

	// 清理和格式化描述
	description = strings.TrimSpace(description)
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

	return description
}
