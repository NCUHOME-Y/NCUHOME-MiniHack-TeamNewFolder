package handle

import (
	"Backend/storage"
	"Backend/trans"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// currentUserID 从 context 里取出登录用户的 ID
// 中间件 auth.Auth（auth/auth.go:24）校验令牌通过后会执行 c.Set("userID", userID)
// 返回值第二个是 bool，表示 "取到了 / 没取到"
func currentUserID(c *gin.Context) (uint64, bool) {
	v, exist := c.Get("userID")
	if !exist {
		return 0, false
	}
	id, ok := v.(uint64)
	if !ok {
		return 0, false
	}
	return id, true
}

// parseID 把路径参数 :id 从字符串转成数字
func parseID(c *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return id, true
}

// ApiCreateComment  POST /api/update/comments
// 发表评论需要登录,挂在 protected 路由组下
func ApiCreateComment(c *gin.Context, s *storage.Store) {
	var req trans.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常，请重新登录"})
		return
	}

	// 逐张校验图片地址：只接受 http(s) 绝对地址或站内绝对路径（如 /uploads/...），
	// 挡住 javascript:、data: 这类会被前端渲染执行的协议（存储型 XSS）
	for _, u := range req.ImageURLs {
		if !isSafeURL(u) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片地址不合法"})
			return
		}
	}

	// 第一张图顺便当封面写进 comments.image_url，
	// 这样评论列表不用再去查 post_images 就能显示一张缩略图
	cover := ""
	if len(req.ImageURLs) > 0 {
		cover = req.ImageURLs[0]
	}

	cm, err := s.CreateComment(c.Request.Context(), req.PostID, userID, req.Content, cover, req.ReplyTo)
		if err != nil {
			// 帖子不存在是客户端的问题，返回 404
			if errors.Is(err, storage.ErrPostNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
				return 
			}
			// 错误详情只写日志，不回显给客户端
			log.Printf("create comment failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败，稍后再试"})
			return
		}
	// 评论本身已经发出去了，只是图片没挂上，记日志即可，不回滚评论
	if len(req.ImageURLs) > 0 {
		if err := s.AddImages(c.Request.Context(), storage.TargetTypeComment, cm.ID, req.ImageURLs); err != nil{
			log.Printf("add comment images failed: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评论成功",
		"comment_id": cm.ID,
		"floor": cm.Floor,
	})
}
// ApiGetCommentList  GET /api/posts/:id/comments
// 公开接口，不需要登录（和 GET /api/posts 一样）
func ApiGetCommentList(c *gin.Context, s *storage.Store) {
	postID, ok := parseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "帖子 ID 不合法"})
		return
	}

	var req trans.CommentListReq
	// GET 请求没有 body，用 ShouldBind 让 Gin 按方法自动选绑定器（GET → query）
	// 千万别用 ShouldBindJSON，空 body 会直接报 EOF，接口恒 400
	// 这个坑你们在 /api/posts 上已经踩过一次了
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：" + err.Error()})
		return
	}

	// 兜底：不传参数时用默认值，也防止有人传 page_size=999
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 50 {
		req.PageSize = 50
	}
	offset := (req.Page - 1) * req.PageSize

	list, total, err := s.GetCommentList(c.Request.Context(), postID, offset, req.PageSize)
	if err != nil {
		log.Printf("get comment list failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评论失败，请稍后重试"})
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

// ApiDeleteComment  DELETE /api/update/comments/:id
// 删除自己的评论。软删除，楼层号保留
func ApiDeleteComment(c *gin.Context, s *storage.Store) {
	commentID, ok := parseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论 ID 不合法"})
		return
	}

	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录状态异常，请重新登录"})
		return
	}

	if err := s.DeleteComment(c.Request.Context(), commentID, userID); err != nil {
		if errors.Is(err, storage.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "评论不存在，或不是你发的"})
			return
		}
		log.Printf("delete comment failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}