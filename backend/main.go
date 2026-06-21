package main

import (
	"log"

	"github.com/IainHenn/chat/redis"
	"github.com/IainHenn/chat/websockets"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := redis.Init(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	r := gin.Default()
	r.GET("/ws", websockets.Ws)
	r.POST("/rooms", redis.CreateRoom)
	r.GET("/rooms", redis.ViewRooms)
	r.Run(":8080")
}
