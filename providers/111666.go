package providers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"imgapi/config"
)

// Provider111666 16图床提供商
type Provider111666 struct{}

// GetName 获取提供商名称
func (p *Provider111666) GetName() string {
	return "111666"
}

// GetDisplayName 获取显示名称
func (p *Provider111666) GetDisplayName() string {
	return "16图床"
}

// IsEnabled 是否启用
func (p *Provider111666) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("111666")
	return exists && providerCfg.Token != ""
}

// Upload 上传图片到16图床
func (p *Provider111666) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("111666")
	if !exists {
		return "", fmt.Errorf("16图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "YOUR-TOKEN-HERE" {
		return "", fmt.Errorf("16图床Token未配置")
	}

	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 创建multipart请求体
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	fileWriter, err := writer.CreateFormFile("image", header.Filename)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return "", fmt.Errorf("复制文件内容失败: %w", err)
	}

	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://i.111666.best/image", &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Auth-Token", providerCfg.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
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

	// 16图床直接返回图片URL
	return string(respBody), nil
}
