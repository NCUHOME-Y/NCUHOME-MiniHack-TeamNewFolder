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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户校验未通过"})
			return
		}

		// 传值：把用户身份塞进 context，交给 handler
		c.Set("userID", userID)

		c.Next()
	}
}
