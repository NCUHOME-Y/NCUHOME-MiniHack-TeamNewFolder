package main

import (
	"Backend/handle"
	"Backend/storage"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "user_1:123456@tcp(127.0.0.1:3306)/demo?parseTime=True"
	db, dberr := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if dberr != nil {
		panic("链接数据库失败")
	}
	s := storage.NewStore(db)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/users/register", func(c *gin.Context) { handle.UsersRegister(c, s) })

		api.POST("/users/login", func(c *gin.Context) { handle.UserLogin(c, s) })
	}

	if err := r.Run("127.0.0.1:8080"); err != nil {
		log.Fatal(err)
	}
}
