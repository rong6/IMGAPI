package providers

import (
	"fmt"
	"mime/multipart"

	"imgapi/config"
)

// ProviderExample 示例图床提供商
// 这是一个示例文件，展示如何实现新的图床提供商
type ProviderExample struct{}

// GetName 获取提供商名称
// 返回的名称应该与config.yaml中的键名一致
func (p *ProviderExample) GetName() string {
	return "example"
}

// GetDisplayName 获取显示名称
// 返回用于展示的友好名称
func (p *ProviderExample) GetDisplayName() string {
	return "示例图床"
}

// IsEnabled 是否启用
// 检查配置文件中是否启用了此提供商，以及必要的配置项是否完整
func (p *ProviderExample) IsEnabled() bool {
	_, exists := config.GetProvider("example")
	if !exists {
		return false
	}

	// 根据实际需要检查必要的配置项
	// 例如：检查token是否配置
	// providerCfg, exists := config.GetProvider("example")
	// return exists && providerCfg.Token != "" && providerCfg.Token != "YOUR-TOKEN-HERE"

	return true
}

// Upload 上传图片
// 实现具体的上传逻辑
func (p *ProviderExample) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 1. 获取配置
	providerCfg, exists := config.GetProvider("example")
	if !exists {
		return "", fmt.Errorf("示例图床未配置")
	}

	// 2. 验证必要的配置项
	if providerCfg.Token == "" || providerCfg.Token == "YOUR-TOKEN-HERE" {
		return "", fmt.Errorf("示例图床Token未配置")
	}

	// 3. 重置文件指针（重要！）
	// file.Seek(0, io.SeekStart)

	// 4. 实现上传逻辑
	// 这里是示例代码，实际需要根据图床API文档实现
	/*
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

		// 添加其他必需字段
		writer.WriteField("key", providerCfg.Token)

		writer.Close()

		// 创建HTTP请求
		req, err := http.NewRequest("POST", "https://example.com/upload", &buf)
		if err != nil {
			return "", fmt.Errorf("创建请求失败: %w", err)
		}

		// 设置请求头
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+providerCfg.Token)

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

		// 解析响应获取图片URL
		// 根据实际API响应格式进行解析
	*/

	// 示例返回
	return "", fmt.Errorf("这是一个示例提供商，不执行实际上传")
}

/*
使用说明：

1. 复制此文件并重命名，例如：新图床.go
2. 修改结构体名称，例如：ProviderNewImageHost
3. 修改GetName()返回的名称，确保与config.yaml中的键名一致
4. 修改GetDisplayName()返回的显示名称
5. 根据需要修改IsEnabled()的逻辑，检查必要的配置项
6. 实现Upload()方法，参考其他提供商的实现
7. 在registry.go中注册新的提供商

配置文件示例：
在config.yaml的providers部分添加：

new_image_host:
  enabled: true
  token: "your-token-here"
  # 其他必要的配置项

常用的HTTP客户端代码模式：

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"mime/multipart"
)

// 创建multipart表单
var buf bytes.Buffer
writer := multipart.NewWriter(&buf)

// 添加文件
fileWriter, _ := writer.CreateFormFile("file", header.Filename)
io.Copy(fileWriter, file)

// 添加其他字段
writer.WriteField("key", "value")
writer.Close()

// 创建请求
req, _ := http.NewRequest("POST", "https://api.example.com/upload", &buf)
req.Header.Set("Content-Type", writer.FormDataContentType())

// 发送请求
client := &http.Client{}
resp, _ := client.Do(req)
defer resp.Body.Close()

// 处理响应
respBody, _ := io.ReadAll(resp.Body)
*/
