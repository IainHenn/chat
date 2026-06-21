package websockets

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // loosen this in production
	},
}

func Ws(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("Failed to connect server via websockets!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect server via websockets!"})
		return
	}

	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()

		if err != nil {
			break
		}

		conn.WriteMessage(msgType, msg)
	}
}
