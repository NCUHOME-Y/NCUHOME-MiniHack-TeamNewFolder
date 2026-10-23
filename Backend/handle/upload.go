package handle

import (
	"Backend/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApiUploadImage  POST /api/update/upload（注意：挂在 /api/update 组下，不是 /api/upload）
// 前端必须用 multipart/form-data 传，字段名固定叫 file
// 返回 {"message":"上传成功","url":"/uploads/2026/10/23/xxx.jpg"}
// 前端拿到 url 后，再把它放进发帖或评论的 image_urls 数组里提交 —— 分两步
func ApiUploadImage(c *gin.Context) {
	// 上传必须登录，否则任何人都能往服务器写文件
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
		return
	}

	url, err := storage.SaveImage(fh)
	if err != nil {
		// 记日志时带上用户 ID，方便追溯谁在传奇怪的东西
		log.Printf("upload failed (user %d): %v", userID, err)
		// 这里的错误文案是我们自己生成的（"图片不能超过 5 MB" 之类），不含敏感信息，可以回显
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "上传成功", "url": url})
}