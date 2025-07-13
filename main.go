package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"imgapi/config"
	"imgapi/handlers"
	"imgapi/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取配置文件路径
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 确保配置文件路径是绝对路径
	if !filepath.IsAbs(configPath) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("获取工作目录失败: %v", err)
		}
		configPath = filepath.Join(wd, configPath)
	}

	// 加载配置
	if err := config.Load(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 启动配置文件监听
	if err := config.WatchConfig(configPath); err != nil {
		log.Printf("启动配置文件监听失败: %v", err)
	}

	// 获取配置
	cfg := config.Get()
	if cfg == nil {
		log.Fatal("配置未加载")
	}

	// 设置Gin模式
	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 注册不需要认证的路由
	r.GET("/getinfo", handlers.GetInfo)

	// 注册需要认证的路由组
	uploadGroup := r.Group("/upload")
	uploadGroup.Use(middleware.AuthMiddleware())
	uploadGroup.POST("/:provider", handlers.Upload)

	// 根路径返回404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"cdoe":    "404",
			"success": false,
			"error":   "页面未找到",
		})
	})

	// 启动服务器
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("聚合图床API启动成功，监听端口: %d", cfg.Server.Port)
	log.Printf("访问 http://localhost%s/getinfo 查看系统信息", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
