package providers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"imgapi/config"
)

// ProviderNoCode NoCode图床提供商
type ProviderNoCode struct{}

// NoCodeResponse NoCode上传响应结构体
type NoCodeResponse struct {
	URL      string      `json:"url"`
	HitRisk  interface{} `json:"hitRisk"`
	RiskTips interface{} `json:"riskTips"`
}

// GetName 获取提供商名称
func (p *ProviderNoCode) GetName() string {
	return "nocode"
}

// GetDisplayName 获取显示名称
func (p *ProviderNoCode) GetDisplayName() string {
	return "NoCode图床"
}

// IsEnabled 是否启用
func (p *ProviderNoCode) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("nocode")
	return exists && providerCfg.Token != ""
}

// Upload 上传图片到NoCode图床
func (p *ProviderNoCode) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("nocode")
	if !exists {
		return "", fmt.Errorf("NoCode图床未配置")
	}

	if providerCfg.Token == "" {
		return "", fmt.Errorf("NoCode图床Cookie未配置，请在config.yaml中设置cookie")
	}

	// 重置文件指针并读取文件数据
	file.Seek(0, io.SeekStart)
	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 检测文件类型
	contentType := http.DetectContentType(fileData)
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("文件类型不支持，仅支持图片格式")
	}

	// 将文件数据转换为base64编码的data URL格式
	base64Data := base64.StdEncoding.EncodeToString(fileData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64Data)

	// 从data URL中提取base64数据部分（逗号后面的部分）
	parts := strings.Split(dataURL, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("生成data URL失败")
	}
	base64Content := parts[1]

	// 构造上传URL
	uploadURL := fmt.Sprintf("https://nocode.cn/api/s3/upload-image?fileName=%s&fileType=%s",
		url.QueryEscape(header.Filename),
		url.QueryEscape(contentType))

	// 创建请求
	req, err := http.NewRequest("POST", uploadURL, strings.NewReader(base64Content))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Cookie", providerCfg.Token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://nocode.cn/")
	req.Header.Set("Origin", "https://nocode.cn")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上传失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var uploadResponse NoCodeResponse
	if err := json.Unmarshal(respBody, &uploadResponse); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if uploadResponse.URL == "" {
		return "", fmt.Errorf("上传失败: 未获得图片URL")
	}

	return uploadResponse.URL, nil
}
