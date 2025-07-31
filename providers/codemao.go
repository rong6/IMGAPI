package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// ProviderCodemao 编程猫图床提供商
type ProviderCodemao struct{}

// CodemaoTokenResponse 编程猫获取Token响应结构体
type CodemaoTokenResponse struct {
	BucketURL string `json:"bucket_url"`
	Data      []struct {
		Token    string `json:"token"`
		Filename string `json:"filename"`
	} `json:"data"`
}

// CodemaoUploadResponse 编程猫上传响应结构体
type CodemaoUploadResponse struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

// GetName 获取提供商名称
func (p *ProviderCodemao) GetName() string {
	return "codemao"
}

// GetDisplayName 获取显示名称
func (p *ProviderCodemao) GetDisplayName() string {
	return "编程猫图床"
}

// IsEnabled 是否启用
func (p *ProviderCodemao) IsEnabled() bool {
	// 无需配置，直接可用
	return true
}

// Upload 上传图片到编程猫图床
func (p *ProviderCodemao) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 读取文件数据
	fileData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 第一步：获取上传Token
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	tokenURL := fmt.Sprintf("https://oversea-api.code.game/tiger/kitten/cdn/token/1?type=%s&prefix=coco%%2Fplayer%%2Funstable&bucket=static",
		url.QueryEscape(contentType))

	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建Token请求失败: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Token请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取Token失败，状态码: %d", resp.StatusCode)
	}

	var tokenResp CodemaoTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("解析Token响应失败: %w", err)
	}

	if len(tokenResp.Data) == 0 {
		return "", fmt.Errorf("获取Token失败: 响应数据为空")
	}

	token := tokenResp.Data[0].Token
	filename := tokenResp.Data[0].Filename

	// 第二步：上传文件到七牛云
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加token字段
	writer.WriteField("token", token)
	// 添加key字段
	writer.WriteField("key", filename)

	// 添加文件字段
	part, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		return "", fmt.Errorf("创建文件字段失败: %w", err)
	}
	part.Write(fileData)

	writer.Close()

	uploadReq, err := http.NewRequest("POST", "https://upload.qiniup.com/", &buf)
	if err != nil {
		return "", fmt.Errorf("创建上传请求失败: %w", err)
	}

	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("文件上传失败，状态码: %d", uploadResp.StatusCode)
	}

	var uploadResult CodemaoUploadResponse
	if err := json.NewDecoder(uploadResp.Body).Decode(&uploadResult); err != nil {
		return "", fmt.Errorf("解析上传响应失败: %w", err)
	}

	if uploadResult.Key == "" {
		return "", fmt.Errorf("上传失败: 未获得文件Key")
	}

	// 拼接最终URL
	finalURL := tokenResp.BucketURL + "/" + uploadResult.Key

	return finalURL, nil
}
