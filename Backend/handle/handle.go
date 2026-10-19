package handle

import (
	"Backend/storage"
	"Backend/trans"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UsersRegister(c *gin.Context, s *storage.Store) {
	var req trans.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	user, err := s.CreateUser(c.Request.Context(), req.Username, req.Password, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败：" + err.Error()})
		return
	}

	resp := trans.UserToUserResp(user)
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "registerinfo": resp})
}

func UserLogin(c *gin.Context, s *storage.Store) {
	var req trans.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	user, err := s.Login(c.Request.Context(), req.Username, req.Password, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败：" + err.Error()})
		return
	}
	resp := trans.UserToLoginResp(user)
	c.JSON(http.StatusAlreadyReported, gin.H{"message": "登陆成功", "logininfo": resp})
}
