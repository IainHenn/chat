package main

import (
	"log"

	"github.com/IainHenn/chat/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := routes.Init(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	r := gin.Default()
	r.GET("/ws", routes.Ws)
	r.POST("/rooms", routes.CreateRoom)
	r.GET("/rooms", routes.ViewRooms)
	r.PUT("/rooms/join", routes.JoinRoom)
	r.PUT("/rooms/leave", routes.LeaveRoom)
	r.DELETE("/rooms", routes.DeleteRoom)
	r.Run(":8080")
}
