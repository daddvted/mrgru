package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type minion struct {
	Hostname string
	Port     string
	Username string
	Password string
	Term     string
}

func indexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	}
}

func connectHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var login minion
		err := c.BindJSON(&login)
		if err != nil {
			fmt.Printf("bind json error%+v\n", err)
			c.JSON(400, gin.H{
				"msg": fmt.Sprintf("%+v", err),
			})
		} else {
			fmt.Printf("%+v\n", login)
			c.JSON(200, gin.H{
				"hello": "ted",
			})

		}
	}
}
