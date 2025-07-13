package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"imgapi/config"
)

// ProviderMeituan 美团图床提供商
type ProviderMeituan struct{}

// MeituanResponse 美团图床响应结构体
type MeituanResponse struct {
	Success bool `json:"success"`
	Data    struct {
		OriginalLink     string `json:"originalLink"`
		OriginalFileName string `json:"originalFileName"`
	} `json:"data"`
}

// GetName 获取提供商名称
func (p *ProviderMeituan) GetName() string {
	return "meituan"
}

// GetDisplayName 获取显示名称
func (p *ProviderMeituan) GetDisplayName() string {
	return "美团图床"
}

// IsEnabled 是否启用
func (p *ProviderMeituan) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("meituan")
	return exists && providerCfg.Token != "" && providerCfg.Token != "换自己的toekn"
}

// Upload 上传图片到美团图床
func (p *ProviderMeituan) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("meituan")
	if !exists {
		return "", fmt.Errorf("美团图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "换自己的toekn" {
		return "", fmt.Errorf("美团图床Token未配置，请在config.yaml中设置token")
	}

	file.Seek(0, io.SeekStart)

	// 创建临时文件保存上传的文件
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("meituan_upload_%d_%s", time.Now().UnixNano(), header.Filename))

	// 创建临时文件
	outFile, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer func() {
		outFile.Close()
		os.Remove(tempFile) // 清理临时文件
	}()

	// 复制文件内容到临时文件
	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	outFile.Close()

	// 使用curl命令上传文件，不然没法成功上传
	cmd := exec.Command("curl", "-s",
		"-X", "POST",
		"https://pic-up.meituan.com/extrastorage/new/video?isHttps=true",
		"-H", "Accept: */*",
		"-H", "Accept-Encoding: gzip, deflate, br",
		"-H", "Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"-H", "Cache-Control: no-cache",
		"-H", "Connection: keep-alive",
		"-H", "Content-Type: multipart/form-data",
		"-H", "Host: pic-up.meituan.com",
		"-H", "Origin: https://czz.meituan.com",
		"-H", "Pragma: no-cache",
		"-H", "Referer: https://czz.meituan.com/",
		"-H", "Sec-Fetch-Dest: empty",
		"-H", "Sec-Fetch-Mode: cors",
		"-H", "Sec-Fetch-Site: same-site",
		"-H", "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0",
		"-H", "client-id: p5gfsvmw6qnwc45n000000000025bbf1",
		"-H", "sec-ch-ua: \"Not A(Brand\";v=\"99\", \"Microsoft Edge\";v=\"121\", \"Chromium\";v=\"121\"",
		"-H", "sec-ch-ua-mobile: ?0",
		"-H", "sec-ch-ua-platform: \"Windows\"",
		"-H", fmt.Sprintf("token: %s", providerCfg.Token),
		"-F", fmt.Sprintf("file=@%s", tempFile),
	)

	// 执行curl命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("curl执行失败: %w, 输出: %s", err, string(output))
	}

	// 查找JSON响应的开始位置
	responseStr := string(output)
	jsonStart := strings.Index(responseStr, "{")
	if jsonStart == -1 {
		return "", fmt.Errorf("未找到JSON响应: %s", responseStr)
	}

	// 提取JSON部分
	jsonResponse := responseStr[jsonStart:]

	// 解析JSON响应
	var meituanResp MeituanResponse
	if err := json.Unmarshal([]byte(jsonResponse), &meituanResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, JSON响应: %s", err, jsonResponse)
	}

	if !meituanResp.Success {
		return "", fmt.Errorf("上传失败，服务器返回: %s", jsonResponse)
	}

	// 返回图片链接
	return meituanResp.Data.OriginalLink, nil
}
