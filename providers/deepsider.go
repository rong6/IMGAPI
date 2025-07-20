package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"imgapi/config"
)

// ProviderDeepSider DeepSider图床提供商
type ProviderDeepSider struct{}

// DeepSiderResponse DeepSider获取签名URL的响应结构体
type DeepSiderResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Data      struct {
		SignedUrl string `json:"signedUrl"`
		Key       string `json:"key"`
		Host      string `json:"host"`
		Filename  string `json:"filename"`
	} `json:"data"`
}

// GetName 获取提供商名称
func (p *ProviderDeepSider) GetName() string {
	return "deepsider"
}

// GetDisplayName 获取显示名称
func (p *ProviderDeepSider) GetDisplayName() string {
	return "DeepSider图床"
}

// IsEnabled 是否启用
func (p *ProviderDeepSider) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("deepsider")
	return exists && providerCfg.Token != "" && providerCfg.Token != "your_token_here"
}

// Upload 上传图片到DeepSider图床
func (p *ProviderDeepSider) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("deepsider")
	if !exists {
		return "", fmt.Errorf("DeepSider图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "your_token_here" {
		return "", fmt.Errorf("DeepSider图床Token未配置，请在config.yaml中设置token（Bearer令牌）")
	}

	// 第一步：获取签名URL
	requestBody := map[string]interface{}{
		"type":     "user-chat",
		"filename": header.Filename,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("构造请求数据失败: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.chargpt.ai/api/upload/private-token", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+providerCfg.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取签名URL失败，状态码: %d", resp.StatusCode)
	}

	var deepSiderResp DeepSiderResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepSiderResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if deepSiderResp.Code != 0 {
		return "", fmt.Errorf("获取签名URL失败: %s", deepSiderResp.Message)
	}

	// 第二步：上传文件到签名URL
	file.Seek(0, io.SeekStart)
	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	uploadReq, err := http.NewRequest("PUT", deepSiderResp.Data.SignedUrl, bytes.NewReader(fileData))
	if err != nil {
		return "", fmt.Errorf("创建上传请求失败: %w", err)
	}

	uploadReq.Header.Set("Content-Type", header.Header.Get("Content-Type"))

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("文件上传失败，状态码: %d", uploadResp.StatusCode)
	}

	// 第三步：拼接最终URL
	host := strings.TrimSuffix(deepSiderResp.Data.Host, "/")
	finalURL := host + "/" + deepSiderResp.Data.Key

	return finalURL, nil
}
