package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

var ws = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := ws.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Printf("fail to upgrade ws:%+v\n", err)
		}

		for {
			t, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(t, msg)
		}

	}
}
