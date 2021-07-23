package web

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func initEngine() *gin.Engine {
	r := gin.Default()
	r.Static("/static", "static")
	r.LoadHTMLGlob("templates/*")

	r.GET("/", indexHandler())
	r.POST("/", connectHandler())
	r.GET("/ws", wsHandler())

	return r
}

func RunHTTPServer(ctx context.Context, mode string) {
	log.Info("Running HTTP server...")
	router := initEngine()

	srv := http.Server{Addr: ":8006", Handler: router}

	go func() {
		<-ctx.Done()
		log.Info("Shutting down HTTP server ...")
		srv.Shutdown(ctx)
	}()
	srv.ListenAndServe()
}
