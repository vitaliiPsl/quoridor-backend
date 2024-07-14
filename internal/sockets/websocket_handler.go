package sockets

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebsocketHandler interface {
	HandleWs(c *gin.Context)
}

type WebsocketHandlerImpl struct {
	service WebsocketService
}

func NewWebsocketHandler(service WebsocketService) *WebsocketHandlerImpl {
	return &WebsocketHandlerImpl{
		service: service,
	}
}

// https://quoridory.domain.io/v1/ws?user_id=1234
func (handler *WebsocketHandlerImpl) HandleWs(c *gin.Context) {
	userId := c.Query("user_id")
	if userId == "" {
		http.Error(c.Writer, "User id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to websocket:", err)
		return
	}

	client := NewWebsocketClient(userId, conn, handler.service)

	handler.service.RegisterClient(client)

	go client.ReadMessage()
	go client.WriteMessage()
}
