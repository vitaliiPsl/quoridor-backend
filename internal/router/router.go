package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Router interface {
	Handler() http.Handler
}

type RouterImpl struct {
	Engine *gin.Engine
}

func NewRouter() *RouterImpl {
	router := gin.Default()

	v1 := router.Group("v1")
	v1.GET("health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "Ok"})
	})

	return &RouterImpl{Engine: router}
}

func (r *RouterImpl) Handler() http.Handler {
	return r.Engine
}
