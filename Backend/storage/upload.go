package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"github.com/google/uuid"
)

const (
	MaxImageSize = 5 << 20 // 5MB
	UploadDir    = "uploads"
	URLPrefix    = "/uploads"
)

var ErrBadImageType = errors.New("只支持 jpg / png / gif / webp格式")


var imageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// SaveImage 接收上传的文件，校验通过后落盘，返回可以访问的站内地址
func SaveImage(fh *multipart.FileHeader) (string, error) {
	if fh == nil {
		return "",errors.New("没有收到文件")
	}
	if fh.Size > MaxImageSize {
		return "", fmt.Errorf("图片不能超过 %d MB", MaxImageSize>>20)
	}

	src, err := fh.Open()
	if err != nil {
		return "", err 
	}
	defer src.Close()// defer 保证函数结束时一定关闭文件，不会泄漏句柄
	
	// 判断类型只需要开头 512 字节
	buf := make([]byte, 512)
	n, _ := src.Read(buf)

	// 刚才 Read 把读指针往前挪了 512 字节，现在必须拨回文件开头，
	// 否则 io.Copy 复制出去的是「从第 512 字节开始」的内容，存下来的图片会损坏、打不开。
	// 注意这里必须是 io.SeekStart（绝对定位到开头），
	// 写成 io.SeekCurrent 表示「从当前位置再偏移 0 字节」，等于没动，图片依然会损坏。
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", err 
	}

	ext, ok := imageExt[http.DetectContentType(buf[:n])]
	if !ok {
		return "", ErrBadImageType
	}

	// 按日期分目录：uploads/2026/10/23/
	dateDir := time.Now().Format("2006/01/02")
	dir := filepath.Join(UploadDir, dateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err 
	}

	// uuid 保证不会重名，也就不会覆盖别人的文件
	name := uuid.NewString() + ext
	dst := filepath.Join(dir, name)

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		os.Remove(dst)
		return "", err
	}

	// filepath.ToSlash 把 Windows 的 \ 换成 /，URL 里必须用 /
	return URLPrefix + "/" + filepath.ToSlash(filepath.Join(dateDir, name)), nil 
}