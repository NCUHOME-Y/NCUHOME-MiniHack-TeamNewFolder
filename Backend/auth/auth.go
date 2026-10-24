package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Store struct {
	db *sql.DB
}

// 中间件验证登录
// 所有需要已登录的操作，要经过此中间件验证Token，并把userID进行传递
func Auth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Token") //抓用户数据
		userID, err := VerifyToken(db, token)
		if err != nil {
			// 【P0 修复】必须用 AbortWithStatusJSON，不能用 c.JSON。
			// 原因：c.JSON 只负责写响应，不会中断 Gin 的中间件链；
			// 后面那个 return 只是退出当前这个闭包，Gin 仍会继续执行真正的业务 handler。
			// 后果有两个：
			//   1) 响应体变成 「{"error":...}{"message":"发布成功"}」两个 JSON 拼接，
			//      前端 JSON.parse 直接抛异常，看起来像后端崩了；
			//   2) 业务逻辑真的跑了一遍 —— 未登录也能写库（越权）。
			// AbortWithStatusJSON = 写响应 + 设置 Abort 标记，后续 handler 才会被跳过。
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户校验未通过"})
			return
		}

		// 传值：把用户身份塞进 context，交给 handler
		c.Set("userID", userID)

		c.Next()
	}
}
