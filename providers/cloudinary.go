package providers

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"imgapi/config"
)

// ProviderCloudinary Cloudinary图床提供商
type ProviderCloudinary struct{}

// GetName 获取提供商名称
func (p *ProviderCloudinary) GetName() string {
	return "cloudinary"
}

// GetDisplayName 获取显示名称
func (p *ProviderCloudinary) GetDisplayName() string {
	return "Cloudinary"
}

// IsEnabled 是否启用
func (p *ProviderCloudinary) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("cloudinary")
	return exists && providerCfg.CloudName != "" && providerCfg.APIKey != "" && providerCfg.APISecret != ""
}

// generateSignature 生成Cloudinary签名
func (p *ProviderCloudinary) generateSignature(params map[string]string, apiSecret string) string {
	// 删除不参与签名的参数
	delete(params, "file")
	delete(params, "api_key")
	delete(params, "resource_type")
	delete(params, "cloud_name")

	// 按键名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signatureItems []string
	for _, k := range keys {
		if params[k] != "" {
			signatureItems = append(signatureItems, fmt.Sprintf("%s=%s", k, params[k]))
		}
	}

	signatureString := strings.Join(signatureItems, "&") + apiSecret

	// 计算SHA1哈希
	h := sha1.New()
	h.Write([]byte(signatureString))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Upload 上传图片到Cloudinary
func (p *ProviderCloudinary) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("cloudinary")
	if !exists {
		return "", fmt.Errorf("Cloudinary未配置")
	}

	if providerCfg.CloudName == "" || providerCfg.CloudName == "your-cloud-name" {
		return "", fmt.Errorf("cloudinary Cloud Name未配置")
	}

	if providerCfg.APIKey == "" || providerCfg.APIKey == "your-api-key" {
		return "", fmt.Errorf("cloudinary API Key未配置")
	}

	if providerCfg.APISecret == "" || providerCfg.APISecret == "your-api-secret" {
		return "", fmt.Errorf("cloudinary API Secret未配置")
	}

	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 当前时间戳
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// 准备签名参数
	signParams := map[string]string{
		"timestamp": timestamp,
	}

	// 生成签名
	signature := p.generateSignature(signParams, providerCfg.APISecret)

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

	// 添加其他字段
	writer.WriteField("api_key", providerCfg.APIKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", signature)

	writer.Close()

	// 构建上传URL
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", providerCfg.CloudName)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", uploadURL, &buf)
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

	// 解析响应获取URL
	// Cloudinary返回的是JSON，需要解析secure_url字段
	// 这里简化处理，实际应该用json.Unmarshal
	respStr := string(respBody)
	if strings.Contains(respStr, "secure_url") {
		// 提取secure_url
		start := strings.Index(respStr, `"secure_url":"`) + 14
		if start > 13 {
			end := strings.Index(respStr[start:], `"`)
			if end > 0 {
				url := respStr[start : start+end]
				// 处理转义字符
				url = strings.ReplaceAll(url, "\\/", "/")
				return url, nil
			}
		}
	}

	return "", fmt.Errorf("无法从响应中提取图片URL: %s", respStr)
}
