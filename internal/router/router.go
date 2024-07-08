package router

import (
	"net/http"
	"quoridor/internal/sockets"

	"github.com/gin-gonic/gin"
)

type Router interface {
	Handler() http.Handler
}

type RouterImpl struct {
	Engine *gin.Engine
}

func NewRouter(websocketHander sockets.WebsocketHandler) *RouterImpl {
	router := gin.Default()

	v1 := router.Group("v1")
	v1.GET("health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "Ok"})
	})

	v1.GET("ws", websocketHander.HandleWs)

	return &RouterImpl{Engine: router}
}

func (r *RouterImpl) Handler() http.Handler {
	return r.Engine
}
