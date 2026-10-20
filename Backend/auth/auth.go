package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Store struct {
	db *sql.DB
}

func Auth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Token")
		userID, err := VerifyToken(db, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户校验未通过"})
			return
		}

		// 传值：把用户身份塞进 context，交给 handler
		c.Set("userID", userID)

		c.Next()
	}
}
