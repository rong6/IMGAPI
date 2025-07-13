package middleware

import (
	"net/http"

	"imgapi/config"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware API密钥验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		if cfg == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "系统配置未加载",
			})
			c.Abort()
			return
		}

		// 如果没有配置key，则跳过验证
		if cfg.API.Key == "" {
			c.Next()
			return
		}

		// 从请求中获取key
		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.Query("key")
		}
		if key == "" {
			key = c.PostForm("key")
		}

		// 验证key
		if key != cfg.API.Key {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "无效的API密钥",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
