package handlers

import (
	"fmt"
	"net/http"
	"time"

	"imgapi/models"
	"imgapi/providers"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// GetInfo 获取系统信息
func GetInfo(c *gin.Context) {
	// 计算运行时间
	uptime := time.Since(startTime)

	// 获取所有提供商信息
	var providerInfos []models.ProviderInfo
	allProviders := providers.GetAllProviders()

	for _, provider := range allProviders {
		providerInfos = append(providerInfos, models.ProviderInfo{
			Name:        provider.GetName(),
			DisplayName: provider.GetDisplayName(),
			Enabled:     provider.IsEnabled(),
			Endpoint:    fmt.Sprintf("/upload/%s", provider.GetName()),
		})
	}

	response := models.InfoResponse{
		Status:    "running",
		Uptime:    uptime.String(),
		Version:   "1.0.0",
		Providers: providerInfos,
	}

	c.JSON(http.StatusOK, response)
}

// Upload 上传图片
func Upload(c *gin.Context) {
	// 获取提供商名称
	providerName := c.Param("provider")

	// 获取提供商
	provider, exists := providers.GetProvider(providerName)
	if !exists {
		c.JSON(http.StatusNotFound, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("未找到图床提供商: %s", providerName),
		})
		return
	}

	// 检查提供商是否启用
	if !provider.IsEnabled() {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("图床提供商 %s 未启用或配置不完整", provider.GetDisplayName()),
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		// 尝试其他常见的字段名
		file, header, err = c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, models.UploadResponse{
				Success: false,
				Error:   "未找到上传文件，请确保字段名为 'image' 或 'file'",
			})
			return
		}
	}
	defer file.Close()

	// 检查文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType != "" && !isImageContentType(contentType) {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   "不支持的文件类型，仅支持图片文件",
		})
		return
	}

	// 检查文件大小（限制为10MB）
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, models.UploadResponse{
			Success: false,
			Error:   "文件大小超过限制（最大10MB）",
		})
		return
	}

	// 上传文件
	url, err := provider.Upload(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("上传失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.UploadResponse{
		Success: true,
		Message: "上传成功",
		URL:     url,
	})
}

// isImageContentType 检查是否为图片类型
func isImageContentType(contentType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
		"image/svg+xml",
	}

	for _, imageType := range imageTypes {
		if contentType == imageType {
			return true
		}
	}
	return false
}
