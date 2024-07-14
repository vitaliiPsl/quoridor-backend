package sockets

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	userId   string
	conn     *websocket.Conn
	service  WebsocketService
	messages chan *WebsocketMessage
}

func NewWebsocketClient(userId string, conn *websocket.Conn, service WebsocketService) *Client {
	return &Client{
		userId:   userId,
		conn:     conn,
		service:  service,
		messages: make(chan *WebsocketMessage, 8),
	}
}

func (c *Client) ReadMessage() {
	log.Printf("Reading messages from the webscoket client. userId=%v...", c.userId)
	defer c.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Error while reading websocket message. Err=%v\n", err)
			break
		}

		wsMessage := WebsocketMessage{}
		if err = json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Error while unmarshaling websocket message. Err=%v\n", err)
			break
		}

		log.Printf("Received webscoket message: userId=%v", c.userId)
		c.service.HandleMessage(c.userId, &wsMessage)
	}
}

func (c *Client) WriteMessage() {
	log.Printf("Writing messsages to the websocket client. userId=%v...", c.userId)
	defer c.Close()

	for message := range c.messages {
		if err := c.conn.WriteJSON(message); err != nil {
			log.Printf("Error while sending message. Err=%v\n", err)
			break
		}
	}
}

func (c *Client) Close() {
	log.Printf("Closing websocket client. userId=%v...", c.userId)
	c.conn.Close()
	c.service.UnregisterClient(c.userId)
}
