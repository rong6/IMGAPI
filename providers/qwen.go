package providers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"imgapi/config"
)

// ProviderQwen Qwen图床提供商
type ProviderQwen struct{}

// QwenTokenResponse Qwen获取Token响应结构体
type QwenTokenResponse struct {
	Success   bool   `json:"success"`
	RequestID string `json:"request_id"`
	Data      struct {
		AccessKeyID     string `json:"access_key_id"`
		AccessKeySecret string `json:"access_key_secret"`
		SecurityToken   string `json:"security_token"`
		FileURL         string `json:"file_url"`
		FilePath        string `json:"file_path"`
		FileID          string `json:"file_id"`
		BucketName      string `json:"bucketname"`
		Region          string `json:"region"`
		Endpoint        string `json:"endpoint"`
	} `json:"data"`
}

// GetName 获取提供商名称
func (p *ProviderQwen) GetName() string {
	return "qwen"
}

// GetDisplayName 获取显示名称
func (p *ProviderQwen) GetDisplayName() string {
	return "Qwen图床"
}

// IsEnabled 是否启用
func (p *ProviderQwen) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("qwen")
	return exists && providerCfg.Token != ""
}

// Upload 上传图片到Qwen图床
func (p *ProviderQwen) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("qwen")
	if !exists {
		return "", fmt.Errorf("Qwen图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "your_umidtoken_here" {
		return "", fmt.Errorf("Qwen图床bx-umidtoken未配置，请在config.yaml中设置token（bx-umidtoken值）")
	}

	// 重置文件指针并读取文件信息
	file.Seek(0, io.SeekStart)
	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	fileSize := len(fileData)

	// 获取Token
	tokenData := map[string]interface{}{
		"filename": header.Filename,
		"filesize": fileSize,
		"filetype": "image",
	}

	jsonData, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("构造Token请求数据失败: %w", err)
	}

	req, err := http.NewRequest("POST", "https://chat.qwen.ai/api/v2/files/getstsToken", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建Token请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://chat.qwen.ai/")
	req.Header.Set("bx-umidtoken", providerCfg.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取Token失败，状态码: %d", resp.StatusCode)
	}

	var tokenResp QwenTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("解析Token响应失败: %w", err)
	}

	if !tokenResp.Success {
		return "", fmt.Errorf("获取Token失败")
	}

	// 上传文件到OSS，使用STS临时凭证和签名
	uploadURL := fmt.Sprintf("https://%s.%s/%s", tokenResp.Data.BucketName, tokenResp.Data.Endpoint, tokenResp.Data.FilePath)

	uploadReq, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(fileData))
	if err != nil {
		return "", fmt.Errorf("创建上传请求失败: %w", err)
	}

	// 设置Content-Type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	uploadReq.Header.Set("Content-Type", contentType)

	// 添加OSS请求头
	now := time.Now().UTC()
	dateStr := now.Format("Mon, 02 Jan 2006 15:04:05 GMT")
	uploadReq.Header.Set("Date", dateStr)
	uploadReq.Header.Set("x-oss-security-token", tokenResp.Data.SecurityToken)

	// 计算OSS签名
	signature := p.calculateOSSSignature("PUT", contentType, dateStr, tokenResp.Data.SecurityToken, "/"+tokenResp.Data.BucketName+"/"+tokenResp.Data.FilePath, tokenResp.Data.AccessKeySecret)
	authorization := fmt.Sprintf("OSS %s:%s", tokenResp.Data.AccessKeyID, signature)
	uploadReq.Header.Set("Authorization", authorization)

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	defer uploadResp.Body.Close()

	// 读取响应内容
	respBody, _ := io.ReadAll(uploadResp.Body)

	if uploadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("文件上传失败，状态码: %d, 响应: %s, URL: %s", uploadResp.StatusCode, string(respBody), uploadURL)
	}

	// 返回文件URL
	if tokenResp.Data.FileURL == "" {
		return "", fmt.Errorf("上传失败: 未获得图片URL")
	}

	return tokenResp.Data.FileURL, nil
}

// calculateOSSSignature 计算OSS签名
func (p *ProviderQwen) calculateOSSSignature(method, contentType, date, securityToken, resource, accessKeySecret string) string {
	// OSS签名字符串格式：
	// StringToSign = VERB + "\n" + Content-MD5 + "\n" + Content-Type + "\n" + Date + "\n" + CanonicalizedOSSHeaders + CanonicalizedResource

	// 构造规范化的OSS头部
	canonicalizedOSSHeaders := fmt.Sprintf("x-oss-security-token:%s\n", securityToken)

	// 构造待签名字符串
	stringToSign := fmt.Sprintf("%s\n\n%s\n%s\n%s%s",
		method,
		contentType,
		date,
		canonicalizedOSSHeaders,
		resource)

	// 使用HMAC-SHA1计算签名
	mac := hmac.New(sha1.New, []byte(accessKeySecret))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return signature
}
