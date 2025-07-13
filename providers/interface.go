package providers

import (
	"mime/multipart"
)

// Provider 图床提供商接口
type Provider interface {
	// GetName 获取提供商名称
	GetName() string

	// GetDisplayName 获取显示名称
	GetDisplayName() string

	// Upload 上传图片
	Upload(file multipart.File, header *multipart.FileHeader) (string, error)

	// IsEnabled 是否启用
	IsEnabled() bool
}
