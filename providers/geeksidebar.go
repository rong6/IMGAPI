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

// ProviderGeeksidebar 极客侧边栏图床提供商
type ProviderGeeksidebar struct{}

// GeeksidebarResponse 极客侧边栏响应结构体
type GeeksidebarResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Success bool   `json:"success"`
	Data    []struct {
		Content  string `json:"content"`
		FileType string `json:"file_type"`
		Filename string `json:"filename"`
		URL      string `json:"url"`
		Type     string `json:"type"`
	} `json:"data"`
}

// GetName 获取提供商名称
func (p *ProviderGeeksidebar) GetName() string {
	return "geeksidebar"
}

// GetDisplayName 获取显示名称
func (p *ProviderGeeksidebar) GetDisplayName() string {
	return "极客侧边栏图床"
}

// IsEnabled 是否启用
func (p *ProviderGeeksidebar) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("geeksidebar")
	return exists && providerCfg.Token != "" && providerCfg.Token != "your_token_here"
}

// Upload 上传图片到极客侧边栏图床
func (p *ProviderGeeksidebar) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("geeksidebar")
	if !exists {
		return "", fmt.Errorf("极客侧边栏图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "your_token_here" {
		return "", fmt.Errorf("极客侧边栏图床Token未配置，请在config.yaml中设置token（Bearer令牌）")
	}

	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 创建multipart请求体
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// sessionsId字段
	writer.WriteField("sessionsId", "a3c7f2e1-5b4d-4c8a-9e2f-1d3b7a6e4f29")

	// 添加文件字段
	part, err := writer.CreateFormFile("files", header.Filename)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		return "", fmt.Errorf("写入文件数据失败: %w", err)
	}

	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://api.geeksidebar.com/v1/api/chat/v2UploadFile", &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+providerCfg.Token)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("上传失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var geeksResp GeeksidebarResponse
	if err := json.NewDecoder(resp.Body).Decode(&geeksResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if !geeksResp.Success || geeksResp.Code != 200 {
		return "", fmt.Errorf("上传失败: %s", geeksResp.Message)
	}

	if len(geeksResp.Data) == 0 {
		return "", fmt.Errorf("上传失败: 响应数据为空")
	}

	if geeksResp.Data[0].URL == "" {
		return "", fmt.Errorf("上传失败: 未获得图片URL")
	}

	return geeksResp.Data[0].URL, nil
}
