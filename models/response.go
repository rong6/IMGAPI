package models

// UploadResponse 上传响应结构体
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// InfoResponse 系统信息响应结构体
type InfoResponse struct {
	Status    string         `json:"status"`
	Uptime    string         `json:"uptime"`
	Version   string         `json:"version"`
	Providers []ProviderInfo `json:"providers"`
}

// ProviderInfo 图床提供商信息
type ProviderInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
	Endpoint    string `json:"endpoint"`
}
