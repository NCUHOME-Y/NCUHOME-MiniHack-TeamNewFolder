package main

import (
	"Backend/auth"
	"Backend/handle"
	"Backend/storage"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 默认配置；生产环境用环境变量覆盖，不设也能直接跑，方便本地联调
const (
	defaultDSN  = "user_1:123456@tcp(127.0.0.1:3306)/demo?parseTime=True&loc=Local"
	defaultAddr = "127.0.0.1:8080"
)

// corsMiddleware 处理跨域。前端跑在不同端口时，浏览器会先发 OPTIONS 预检，
// 必须在 Allow-Headers 里放行本项目自定义的 Token 头，否则联调全部被拦。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			// 非浏览器请求（curl / 服务端调用）没有 Origin
			origin = "*"
		}
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Token")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dsn := envOr("DB_DSN", defaultDSN)
	addr := envOr("ADDR", defaultAddr)

	db, dberr := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if dberr != nil {
		// 带上真实错误，否则看不出是密码错、库不存在还是网络不通
		log.Fatalf("连接数据库失败: %v", dberr)
	}

	sqldb, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}
	// 不设连接池默认不限制连接数，并发上来会打满 MySQL
	sqldb.SetMaxOpenConns(50)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxLifetime(30 * time.Minute)

	s := storage.NewStore(db)
	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api")
	{
		api.POST("/users/register", func(c *gin.Context) { handle.ApiUsersRegister(c, s, sqldb) })
		api.POST("/users/login", func(c *gin.Context) { handle.ApiUserLogin(c, s, sqldb) })
		api.GET("/posts", func(c *gin.Context) { handle.ApiGetPostList(c, s) })

		protected := api.Group("/update")
		protected.Use(auth.Auth(sqldb))
		{
			protected.POST("/create/post", func(c *gin.Context) { handle.ApiUserCreatePost(c, s) })
			protected.POST("/bio", func(c *gin.Context) { handle.ApiUserUpdateBio(c, s) })
			protected.POST("/avatar", func(c *gin.Context) { handle.ApiUserUpdateAvatar(c, s) })
			protected.POST("/logout", func(c *gin.Context) { handle.ApiUserLogout(c, s) })
		}
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		log.Printf("服务启动: http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务异常退出: %v", err)
		}
	}()

	// 优雅关闭：收到退出信号后先停止接收新请求，再等待在途请求处理完
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务关闭超时，强制退出: %v", err)
	}
	log.Println("服务已停止")
}
