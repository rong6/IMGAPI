package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"imgapi/config"
)

// ProviderNodeSeek NodeSeek图床提供商
type ProviderNodeSeek struct{}

// NodeSeekResponse NodeSeek图床响应结构体
type NodeSeekResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ImageID  string `json:"image_id"`
	Filename string `json:"filename"`
	Size     int    `json:"size"`
	Links    struct {
		Direct   string `json:"direct"`
		HTML     string `json:"html"`
		Markdown string `json:"markdown"`
		BBCode   string `json:"bbcode"`
	} `json:"links"`
}

// GetName 获取提供商名称
func (p *ProviderNodeSeek) GetName() string {
	return "nodeimage"
}

// GetDisplayName 获取显示名称
func (p *ProviderNodeSeek) GetDisplayName() string {
	return "NodeSeek图床"
}

// IsEnabled 是否启用
func (p *ProviderNodeSeek) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("nodeimage")
	return exists && providerCfg.Token != "" && providerCfg.Token != "your_api_key_here"
}

// Upload 上传图片到NodeSeek图床
func (p *ProviderNodeSeek) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("nodeimage")
	if !exists {
		return "", fmt.Errorf("nodeSeek图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "your_api_key_here" {
		return "", fmt.Errorf("nodeSeek图床API密钥未配置，请在config.yaml中设置token")
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
	req, err := http.NewRequest("POST", "https://api.nodeimage.com/api/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-KEY", providerCfg.Token)

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

	// 解析JSON响应
	var nodeSeekResp NodeSeekResponse
	if err := json.Unmarshal(respBody, &nodeSeekResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, 原始响应: %s", err, string(respBody))
	}

	if !nodeSeekResp.Success {
		return "", fmt.Errorf("上传失败: %s", nodeSeekResp.Message)
	}

	// 移除调试输出
	// fmt.Printf("Debug NodeSeek响应: %s\n", string(respBody))

	// 返回直链URL
	if nodeSeekResp.Links.Direct != "" {
		return nodeSeekResp.Links.Direct, nil
	}

	return "", fmt.Errorf("上传成功但未获得图片URL，响应: %s", string(respBody))
}
