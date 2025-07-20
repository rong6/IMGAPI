package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// ProviderZycsimg 骤雨重山图床提供商
type ProviderZycsimg struct{}

// ZycsimgResponse 骤雨重山图床响应结构体
type ZycsimgResponse struct {
	Status  int  `json:"status"`
	Success bool `json:"success"`
	Data    struct {
		ID         string `json:"id"`
		DeleteHash string `json:"deletehash"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		Size       int    `json:"size"`
		Link       string `json:"link"`
		Datetime   int64  `json:"datetime"`
	} `json:"data"`
}

// GetName 获取提供商名称
func (p *ProviderZycsimg) GetName() string {
	return "zycsimg"
}

// GetDisplayName 获取显示名称
func (p *ProviderZycsimg) GetDisplayName() string {
	return "骤雨重山图床"
}

// IsEnabled 是否启用
func (p *ProviderZycsimg) IsEnabled() bool {
	// 无需配置，直接可用
	return true
}

// Upload 上传图片到骤雨重山图床
func (p *ProviderZycsimg) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
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

	// 添加文件字段
	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		return "", fmt.Errorf("写入文件数据失败: %w", err)
	}

	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://wp-cdn.4ce.cn/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	var zycsResp ZycsimgResponse
	if err := json.NewDecoder(resp.Body).Decode(&zycsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if !zycsResp.Success || zycsResp.Status != 200 {
		return "", fmt.Errorf("上传失败: 状态码 %d", zycsResp.Status)
	}

	if zycsResp.Data.ID == "" {
		return "", fmt.Errorf("上传失败: 未获得图片ID")
	}

	// 拼接最终URL
	finalURL := fmt.Sprintf("https://wp-cdn.4ce.cn/v2/%s.jpeg", zycsResp.Data.ID)

	return finalURL, nil
}
