package main

import (
	"Backend/auth"
	"Backend/handle"
	"Backend/storage"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "user_1:123456@tcp(127.0.0.1:3306)/demo?parseTime=True&loc=Local"
	db, dberr := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if dberr != nil {
		panic("链接数据库失败")
	}

	sqldb, err := db.DB()
	if err != nil {
		panic("数据库处理错误")
	}

	s := storage.NewStore(db)
	r := gin.Default()

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

	if err := r.Run("127.0.0.1:8080"); err != nil {
		log.Fatal(err)
	}
}
