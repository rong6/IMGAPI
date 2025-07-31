<div align="center">
<h1>IMGAPI - 聚合图床API</h1>
<img src="https://socialify.git.ci/rong6/IMGAPI/image?description=1&language=1&font=Inter&name=1&owner=1&pattern=Circuit%20Board&theme=Dark" alt="Cover Image" width="650"><br>
<a href="#快速开始">快速开始</a> | <a href="DOCS.md">配置文档</a>
</div>

## 功能特性

- 🚀 支持多个图床平台
- 🔑 API密钥校验
- ♻️ 配置热加载
- 📝 RESTful API设计
- 🛡️ 安全的文件上传
- 🎯 易于扩展

## 支持的图床

- **16图床 (111666)** - 需要配置
- **美团图床 (meituan)** - 需要配置
- **360图床 (360tu)** - 无需配置
- **Cloudinary (cloudinary)** - 需要配置
- **IPFS图床 (ipfs)** - 无需配置
- **NodeSeek图床 (nodeimage)** - 需要配置
- **EroLabs图床 (erolabs)** - 需要配置
- **DeepSider图床 (deepsider)** - 需要配置
- **骤雨重山图床 (zycsimg)** - 无需配置
- **编程猫图床 (codemao)** - 无需配置
- **极客侧边栏图床 (geeksidebar)** - 需要配置
- **Qwen图床 (qwen)** - 需要配置
- **NoCode图床 (nocode)** - 需要配置

## 快速开始

### 方式一：Docker部署（推荐）

克隆代码并编辑配置：
``` bash
# 克隆项目
git clone https://github.com/rong6/IMGAPI
cd imgapi

# 编辑配置文件
cp config.yaml config.yaml.example
vim config.yaml
```

1. **使用docker-compose（最简单）**

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

2. **直接使用Docker**

```bash
# 构建镜像
docker build -t imgapi .

# 运行容器
docker run -d \
  --name imgapi \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  --restart unless-stopped \
  imgapi

# 查看日志
docker logs -f imgapi
```

### 方式二：源码编译

1. **安装依赖**

确保您的系统已安装Go 1.19或更高版本。

```bash
go mod download
```

2. **编译和运行**

```bash
# 下载依赖
go mod tidy

# 编译
go build -o imgapi

# 运行
./imgapi
```

### 3. 配置

编辑 `config.yaml` 文件即可。对于详细配置参阅[DOCS.md](DOCS.md)。

### 4. 访问服务

服务启动后，访问 `http://localhost:8080/getinfo` 查看是否正常启动。你可以使用项目中的 [demo.html](demo.html) 文件检验API是否正常运作。

## API 接口

### 获取系统信息

**请求**
```
GET /getinfo
```

**响应**
```json
{
  "status": "running",
  "uptime": "1h23m45s",
  "version": "1.0.0",
  "providers": [
    {
      "name": "meituan",
      "display_name": "美团图床",
      "enabled": true,
      "endpoint": "/upload/meituan"
    }
  ]
}
```

### 上传图片

**请求**
```
POST /upload/:provider
Content-Type: multipart/form-data

- image: 图片文件（字段名也可以是 file）
- key: API密钥（如果配置了的话）
```

**示例**
```bash
# 上传到美团图床
curl -F image=@your-image.jpg http://localhost:8080/upload/meituan

# 使用API密钥
curl -F image=@your-image.jpg -F key=your-api-key http://localhost:8080/upload/meituan

# 或者通过Header传递密钥
curl -F image=@your-image.jpg -H "X-API-Key: your-api-key" http://localhost:8080/upload/meituan
```

**响应**
```json
{
  "success": true,
  "message": "上传成功",
  "url": "https://example.com/image.jpg"
}
```

**错误响应**
```json
{
  "success": false,
  "error": "错误描述"
}
```

## 扩展新的图床提供商

1. 在 `providers` 目录下创建新文件，例如 `newprovider.go`
2. 实现 `Provider` 接口：

```go
package providers

import (
    "mime/multipart"
    "imgapi/config"
)

type ProviderNewProvider struct{}

func (p *ProviderNewProvider) GetName() string {
    return "newprovider"
}

func (p *ProviderNewProvider) GetDisplayName() string {
    return "新图床"
}

func (p *ProviderNewProvider) IsEnabled() bool {
    _, exists := config.GetProvider("newprovider")
    return exists
}

func (p *ProviderNewProvider) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
    // 实现上传逻辑
    return "https://example.com/uploaded-image.jpg", nil
}
```

3. 在 `providers/registry.go` 的 `init()` 函数中注册：

```go
func init() {
    // 其他提供商...
    RegisterProvider(&ProviderNewProvider{})
}
```

4. 在 `config.yaml` 中添加配置：

```yaml
providers:
  newprovider:
    enabled: true
    # 其他必要的配置项
```


## 项目结构

```
imgapi/
│  .dockerignore           # Docker构建忽略文件
│  .gitignore             # Git忽略文件
│  config.yaml            # 配置文件
│  demo.html              # 前端测试页面
│  docker-compose.yml     # Docker Compose配置
│  Dockerfile             # Docker构建文件
│  go.mod                 # Go模块文件
│  go.sum                 # Go依赖锁定文件
│  main.go                # 程序入口
│  README.md              # 项目文档
│
├─config/                 # 配置管理
│      config.go
│
├─handlers/               # HTTP处理器
│      handlers.go
│
├─middleware/             # 中间件
│      auth.go
│
├─models/                 # 数据模型
│      response.go
│
└─providers/              # 图床提供商
       111666.go          # 16图床
       360tu.go           # 360图床
       cloudinary.go      # Cloudinary
       codemao.go         # 编程猫图床
       deepsider.go       # DeepSider
       example.go         # 示例实现
       geeksidebar.go     # 极客侧边栏图床
       interface.go       # 提供商接口
       ipfs.go            # IPFS图床
       meituan.go         # 美团图床
       nocode.go          # NoCode图床
       nodeseek.go        # NodeSeek图床
       qwen.go            # QWen图床
       registry.go        # 提供商注册表
       zycsimg.go         # 骤雨重山图床
```

## 安全特性

- 支持API密钥验证
- 文件类型检查（仅允许图片）
- 文件大小限制（默认10MB）
- CORS配置
- 配置文件热加载避免密钥泄露

## 免责声明

本项目仅供学习和研究使用，请勿将其用于任何违法用途。使用本项目时，请注意：

1. **服务可用性**：本项目不保证所集成的第三方图床服务的稳定性和可用性，图床服务可能随时发生变化或停止服务。

2. **数据安全**：上传的图片将存储在第三方图床服务上，请勿上传包含敏感信息、个人隐私或违法内容的图片。

3. **合规使用**：请确保上传的内容符合相关法律法规和第三方平台的服务条款，任何违法违规行为与本项目无关。

4. **版权责任**：用户需确保上传的图片拥有合法的使用权限，因版权纠纷产生的法律责任由用户自行承担。

5. **使用风险**：使用本项目可能存在数据丢失、服务中断等风险，开发者不承担任何直接或间接的损失责任。

6. **第三方服务**：本项目仅作为第三方图床服务的接口聚合工具，对第三方服务的政策变更、服务质量等不承担责任。

使用本项目即表示您已阅读、理解并同意上述免责声明。如不同意，请立即停止使用。

## 许可证

MIT License
