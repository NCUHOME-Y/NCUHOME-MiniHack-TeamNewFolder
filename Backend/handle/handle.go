package handle

import (
	"Backend/Tags"
	"Backend/auth"
	"Backend/storage"
	"Backend/trans"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 默认Token过期时间
const (
	AccessTokenExpire = 4 * time.Hour
)

// bindJSON 解析 JSON 请求体；请求体为空时直接跳过绑定。
// 用于 logout 这类「body 可选」的接口，避免前端不传 body 时被判成 400。
func bindJSON(c *gin.Context, obj any) error {
	if c.Request.ContentLength == 0 {
		return nil
	}
	return c.ShouldBindJSON(obj)
}

// isSafeURL 判断地址是否可以入库：仅允许 http/https 绝对地址或站内绝对路径，
// 用于挡住 javascript:、data: 等会被前端渲染执行的协议（存储型 XSS）。
func isSafeURL(raw string) bool {
	u := strings.TrimSpace(raw)
	if u == "" {
		return false
	}
	if strings.HasPrefix(u, "/") && !strings.HasPrefix(u, "//") {
		return true // 站内绝对路径，如 /uploads/a.png
	}
	lower := strings.ToLower(u)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func ApiUsersRegister(c *gin.Context, s *storage.Store, db *sql.DB) {
	var req trans.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	user, err := s.CreateUser(c.Request.Context(), req.Username, req.Password, "")
	if err != nil {
		// 必须 return，否则会继续往下用 nil user 取 user.ID 并二次写响应
		if errors.Is(err, storage.ErrUserExists) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"code": "USERNAME_TAKEN", "error": "用户名已被占用"})
			return
		}
		log.Printf("register failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

	token, err := auth.GenerateToken(db, user.ID, AccessTokenExpire)
	if err != nil {
		log.Printf("generate token failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token生成失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "registerinfo": token})
}

func ApiUserLogin(c *gin.Context, s *storage.Store, db *sql.DB) {
	var req trans.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	user, err := s.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := auth.GenerateToken(db, user.ID, AccessTokenExpire)
	if err != nil {
		// 这是服务端故障，不是认证失败，应返回 500
		log.Printf("generate token failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Token获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "登陆成功", "token": token})
}

func ApiUserUpdateBio(c *gin.Context, s *storage.Store) {
	var req trans.UpdateBioReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	userID, exist := c.Get("userID")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	id, ok := userID.(uint64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常"})
		return
	}

	err := s.UserUpdateBio(id, req.Bio)
	if err != nil {
		// 错误详情只进日志，不回显给客户端
		log.Printf("update bio failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "数据存储时发生错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func ApiUserUpdateAvatar(c *gin.Context, s *storage.Store) {
	var req trans.UpdateAvatarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	userID, exist := c.Get("userID")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	id, ok := userID.(uint64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常"})
		return
	}
	// 只接受 http/https 绝对地址或站内绝对路径，避免 javascript: 之类被前端渲染执行
	if req.AvatarURL != "" && !isSafeURL(req.AvatarURL) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "头像地址不合法"})
		return
	}

	err := s.UserUpdateAvatar(id, req.AvatarURL)
	if err != nil {
		// 错误详情只进日志，不回显给客户端
		log.Printf("update avatar failed: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "数据存储时发生错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func ApiUserLogout(c *gin.Context, s *storage.Store) {
	var req trans.Token
	// body 可选：不传也能退，此时用请求头里的令牌
	if err := bindJSON(c, &req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	_, exist := c.Get("userID")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	rawToken := req.Token
	if rawToken == "" {
		rawToken = c.GetHeader("Token")
	}
	if rawToken == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "缺少要注销的令牌"})
		return
	}

	err := s.DeleteToken(rawToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "数据库访问失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "退出成功"})
}

func ApiUserCreatePost(c *gin.Context, s *storage.Store) {
	var req trans.CreatePost
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	ID, exist := c.Get("userID")
	if !exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	userID, ok := ID.(uint64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常"})
		return
	}

	// 封面图：优先用 image_url，没传就用 image_urls 里的第一张
	cover := req.ImageURL
	if cover == "" && len(req.ImageURLs) > 0 {
		cover = req.ImageURLs[0]
	}
	// 校验地址合法性：只接受 http(s) 绝对地址或站内绝对路径（如 /uploads/...），
	// 挡住 javascript:、data: 这类会被前端渲染执行的协议（存储型 XSS）
	if cover != "" && !isSafeURL(cover) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "图片地址不合法"})
		return
	}
	for _, u := range req.ImageURLs {
		if !isSafeURL(u) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "图片地址不合法"})
			return
		}
	}

	normalized, err := Tags.NormTag(c.Request.Context(), s.DB, req.Tags)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标签不合法"})
		return
	}

	post := storage.Post{
		UserID:      userID,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    cover,
		Tags:        normalized,
		Location:    req.Location,
		ContactInfo: req.ContactInfo,
		Status:      0,
	}

	err = s.UserCreatePost(c.Request.Context(), &post)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "发布失败，请稍后重试"})
		return
	}

	// 多图写进 post_images。失败只记日志：帖子本身已经创建成功，不该因此回滚
	if len(req.ImageURLs) > 0 {
		if err := s.AddImages(c.Request.Context(), storage.TargetTypePost, post.ID, req.ImageURLs); err != nil {
			log.Printf("add post images failed: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "发布成功", "post_id": post.ID})
}

func ApiGetPostList(c *gin.Context, s *storage.Store) {
	var req trans.PostListReq
	// GET 请求没有 body，用 ShouldBind 会由框架按方法自动选择绑定器：
	// GET → query（form 标签与 default 生效），若以后改成 POST + JSON 也能兼容。
	if err := c.ShouldBind(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	// 兜底，防止 limit/offset 为 0 时 GORM 视作不限制而返回整表
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 50 {
		req.PageSize = 50
	}

	offset := (req.Page - 1) * req.PageSize
	limit := req.PageSize

	var queryType int8 = 0
	if req.Type != nil {
		queryType = *req.Type
	}

	defaultStatus := int8(-1)
	list, total, err := s.GetPostList(c.Request.Context(), queryType, defaultStatus, offset, limit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取成功",
		"data": gin.H{
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"list":      list,
		},
	})
}

// ApiPostContactInfo  GET /api/posts/:id/contactinfo （需登录）
// 返回某个帖子的联系方式。联系方式在列表里故意不返回，只在详情里给，
// 所以这个接口是「给失主/拾到者留联系方式」的唯一出口。
func ApiPostContactInfo(c *gin.Context, s *storage.Store) {
	// 【P0 修复】帖子 ID 从 URL 路径参数里取，不再从请求体读。
	// 原来写的是 ShouldBindJSON(&req) + trans.PostDetailReq，但这里路由是 GET：
	//   - 浏览器 fetch 无法给 GET 请求带 body（前端 README 里也是这么记的）
	//   - JSON 解码器读到空流立刻返回 EOF，于是这个接口 100% 返回 400
	// 也就是说「获取联系方式」这条路一直是断的，详情页只能显示"请到留言区联系"。
	// parseID 是 handle/comment.go:31 里现成的工具函数，和评论接口共用同一套解析逻辑。
	postID, ok := parseID(c, "id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "帖子 ID 不合法"})
		return
	}

	query, err := s.GetPostDetail(c.Request.Context(), postID)
	if err != nil {
		// 区分「客户端问题」和「服务端问题」，不再一股脑 500
		switch {
		case errors.Is(err, storage.ErrPostNotFound):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
		case errors.Is(err, storage.ErrPostClosed):
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "帖子已结帖，联系方式不可见"})
		default:
			// 真实错误只写日志，不回显给客户端（响应里带 SQL 错误会泄露表结构）
			log.Printf("get post detail failed: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "获取帖子详情失败，请稍后重试"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "获取成功",
		"contactinfo": query.ContactInfo,
	})
}

func ApiGetMeInfo(c *gin.Context, s *storage.Store) {
	ID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	userID, ok := ID.(uint64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常"})
		return
	}

	u, err := s.GetUserByID(c, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "获取用户详细信息失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":   u.Username,
		"bio":        u.Bio,
		"avatar_url": u.AvatarURL,
		"create_at":  u.CreatAt,
	})
}
