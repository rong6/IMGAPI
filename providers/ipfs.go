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

// ProviderIPFS IPFS图床提供商
type ProviderIPFS struct{}

// IPFSResponse IPFS图床响应结构体
type IPFSResponse struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
	Size string `json:"Size"`
	URL  string `json:"Url"`
}

// GetName 获取提供商名称
func (p *ProviderIPFS) GetName() string {
	return "ipfs"
}

// GetDisplayName 获取显示名称
func (p *ProviderIPFS) GetDisplayName() string {
	return "IPFS图床"
}

// IsEnabled 是否启用
func (p *ProviderIPFS) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("ipfs")
	return exists && providerCfg.Enabled
}

// Upload 上传图片到IPFS图床
func (p *ProviderIPFS) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 创建multipart请求体
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件字段
	fileWriter, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return "", fmt.Errorf("复制文件内容失败: %w", err)
	}

	writer.Close()

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://api.img2ipfs.org/api/v0/add?pin=false", &buf)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
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

	// 解析JSON响应
	var ipfsResp IPFSResponse
	if err := json.Unmarshal(respBody, &ipfsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if ipfsResp.URL == "" {
		return "", fmt.Errorf("上传失败，未获得图片URL")
	}

	return ipfsResp.URL, nil
}
