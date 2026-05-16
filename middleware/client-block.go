package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

// ClientBlock 中间件：拦截特定客户端的请求
//
// 环境变量配置：
//
//	BLOCK_NODEFETCH=true        — 拦截 User-Agent 包含 node-fetch 的请求（SillyTavern 默认）
//	BLOCKED_USER_AGENTS=xxx,yyy — 额外拦截的 User-Agent 关键词（逗号分隔）
//	REQUIRE_CLIENT_HEADER=X-Custom-Token  — 要求请求必须带此 header
//	REQUIRE_CLIENT_HEADER_VALUE=secret    — 该 header 的值必须匹配（不设置则只检查存在性）
func ClientBlock() gin.HandlerFunc {
	blockNodeFetch := os.Getenv("BLOCK_NODEFETCH") == "true"
	blockedUAs := parseCommaList(os.Getenv("BLOCKED_USER_AGENTS"))
	requireHeader := os.Getenv("REQUIRE_CLIENT_HEADER")
	requireHeaderValue := os.Getenv("REQUIRE_CLIENT_HEADER_VALUE")

	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		userId := c.GetInt("id")

		// 检查 node-fetch（SillyTavern 特征）
		if blockNodeFetch && strings.Contains(strings.ToLower(userAgent), "node-fetch") {
			logger.LogWarn(c.Request.Context(), "client-block: blocked node-fetch request from user %d, UA: %s", userId, userAgent)
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": "该客户端不被允许访问此服务",
					"type":    "client_blocked",
					"code":    "client_not_allowed",
				},
			})
			c.Abort()
			return
		}

		// 检查自定义 UA 黑名单
		for _, blocked := range blockedUAs {
			if blocked != "" && strings.Contains(strings.ToLower(userAgent), strings.ToLower(blocked)) {
				logger.LogWarn(c.Request.Context(), "client-block: blocked request from user %d, matched UA keyword: %s", userId, blocked)
				c.JSON(http.StatusForbidden, gin.H{
					"error": gin.H{
						"message": "该客户端不被允许访问此服务",
						"type":    "client_blocked",
						"code":    "client_not_allowed",
					},
				})
				c.Abort()
				return
			}
		}

		// 检查必需的自定义 header
		if requireHeader != "" {
			headerValue := c.GetHeader(requireHeader)
			if headerValue == "" {
				logger.LogWarn(c.Request.Context(), "client-block: missing required header '%s' from user %d", requireHeader, userId)
				c.JSON(http.StatusForbidden, gin.H{
					"error": gin.H{
						"message": "缺少必要的认证头",
						"type":    "client_blocked",
						"code":    "missing_client_header",
					},
				})
				c.Abort()
				return
			}
			if requireHeaderValue != "" && headerValue != requireHeaderValue {
				logger.LogWarn(c.Request.Context(), "client-block: invalid header value from user %d", userId)
				c.JSON(http.StatusForbidden, gin.H{
					"error": gin.H{
						"message": "客户端认证失败",
						"type":    "client_blocked",
						"code":    "invalid_client_header",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func parseCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
