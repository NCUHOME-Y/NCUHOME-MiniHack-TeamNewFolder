package handle

import (
	"Backend/auth"
	"Backend/storage"
	"Backend/trans"
	"database/sql"
	"errors"
	"log"
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
	switch {
	case errors.Is(err, storage.ErrUserExists):
		c.JSON(http.StatusConflict, gin.H{"code": "USERNAME_TAKEN", "error": "用户名已被占用"})
	default:
		log.Printf("register failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
	}

	token, err := auth.GenerateToken(db, user.ID, AccessTokenExpire)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token生成失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "registerinfo": token})
}

func ApiUserLogin(c *gin.Context, s *storage.Store, db *sql.DB) {
	var req trans.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	user, err := s.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := auth.GenerateToken(db, user.ID, AccessTokenExpire)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token获取失败：" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "登陆成功", "token": token})
}

func ApiUserUpdateBio(c *gin.Context, s *storage.Store) {
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
	id, ok := userID.(uint64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常"})
		return
	}

	err := s.UserUpdateBio(id, req.Bio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据存储时发生错误" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}
func ApiUserUpdateAvatar(c *gin.Context, s *storage.Store) {
	var req trans.UpdateAvatarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	userID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	id, ok := userID.(uint64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常"})
		return
	}
	err := s.UserUpdateAvatar(id, req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据存储时发生错误" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func ApiUserLogout(c *gin.Context, s *storage.Store) {
	var req trans.Token
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	_, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	err := s.DeleteToken(req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库访问失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
}

func ApiUserCreatePost(c *gin.Context, s *storage.Store) {
	var req trans.CreatePost
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	ID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	userID := ID.(uint64)

	post := storage.Post{
		UserID:      userID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Tags:        req.Tags,
		Location:    req.Location,
		ContactInfo: req.ContactInfo,
		Status:      0,
	}

	err := s.UserCreatePost(&post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "post_id": post.ID})
}
