package handle

import (
	"Backend/auth"
	"Backend/storage"
	"Backend/trans"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenExpire = 4 * time.Hour
)

func ApiUsersRegister(c *gin.Context, s *storage.Store, db *sql.DB) {
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

	token, err := auth.GenerateToken(db, user.ID, AccessTokenExpire)
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "registerinfo": token})
}

func ApiUserLogin(c *gin.Context, s *storage.Store, db *sql.DB) {
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

	token, err := auth.GenerateToken(db, user.ID, AccessTokenExpire)
	c.JSON(http.StatusAlreadyReported, gin.H{"message": "登陆成功", "token": token})
}

func ApiUserUpdateBio(c *gin.Context, db *sql.DB) {
	var req trans.UpdateBioReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	userID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	id := userID.(uint64)

	gormDB, err := storage.ConvertToGormDB(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库初始化失败: " + err.Error()})
		return
	}

	s := storage.NewStore(gormDB)
	err = s.UserUpdateBio(id, req.Bio)

}
