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

// ProviderEroLabs EroLabs图床提供商
type ProviderEroLabs struct{}

// EroLabsResponse EroLabs图床响应结构体
type EroLabsResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
	Data      struct {
		URL     string `json:"url"`
		Message string `json:"message"`
	} `json:"data"`
}

// GetName 获取提供商名称
func (p *ProviderEroLabs) GetName() string {
	return "erolabs"
}

// GetDisplayName 获取显示名称
func (p *ProviderEroLabs) GetDisplayName() string {
	return "EroLabs图床"
}

// IsEnabled 是否启用
func (p *ProviderEroLabs) IsEnabled() bool {
	providerCfg, exists := config.GetProvider("erolabs")
	return exists && providerCfg.Token != "" && providerCfg.Token != "your_cookie_here"
}

// Upload 上传图片到EroLabs图床
func (p *ProviderEroLabs) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	providerCfg, exists := config.GetProvider("erolabs")
	if !exists {
		return "", fmt.Errorf("eroLabs图床未配置")
	}

	if providerCfg.Token == "" || providerCfg.Token == "your_cookie_here" {
		return "", fmt.Errorf("eroLabs图床Cookie未配置，请在config.yaml中设置token（Cookie值）")
	}

	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 创建临时文件保存上传的文件
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("erolabs_upload_%d_%s", time.Now().UnixNano(), header.Filename))

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

	// 使用curl命令上传文件
	cmd := exec.Command("curl", "-s",
		"-X", "POST",
		"https://game.ero-labs.gold/api/v2/forum/1/article/cce1efc7-114f-f9e8-2a11-2d2082eb99cd/pic/upload",
		"-H", "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0",
		"-H", "Accept: application/json, text/javascript, */*; q=0.01",
		"-H", fmt.Sprintf("Cookie: %s", providerCfg.Token),
		"-F", fmt.Sprintf("file=@%s;filename=%q;headers=\"Content-Type: %s\"", tempFile, header.Filename, header.Header.Get("Content-Type")),
	)
	// 执行curl命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("curl执行失败: %w, 输出: %s", err, string(output))
	}

	responseStr := string(output)

	// 查找JSON响应
	jsonStart := strings.Index(responseStr, "{")
	if jsonStart == -1 {
		return "", fmt.Errorf("上传失败: 服务器未返回有效响应")
	}

	jsonResponse := responseStr[jsonStart:]

	// 解析JSON响应
	var eroLabsResp EroLabsResponse
	if err := json.Unmarshal([]byte(jsonResponse), &eroLabsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w, JSON响应: %s", err, jsonResponse)
	}

	if eroLabsResp.Status != "SUCCESS" {
		return "", fmt.Errorf("上传失败: %s", eroLabsResp.Message)
	}

	if eroLabsResp.Data.URL == "" {
		return "", fmt.Errorf("上传成功但未获得图片URL，响应: %s", jsonResponse)
	}

	return eroLabsResp.Data.URL, nil
}
