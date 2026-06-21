package redis

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

var Client *goredis.Client

func Init() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	Client = goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	return Client.Ping(context.Background()).Err()
}

type RoomCreationReply struct {
	RoomName string `json:"room_name"`
}

func CreateRoom(c *gin.Context) {
	var reply RoomCreationReply

	if err := c.BindJSON(&reply); err != nil {
		fmt.Println(reply.RoomName)
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create room, invalid body!"})
		return
	}

	var shorthandName string

	if len(reply.RoomName) < 10 {
		shorthandName = reply.RoomName
	} else {
		shorthandName = reply.RoomName[:10]
	}

	roomID := "room:" + shorthandName

	Client.SAdd(c, "rooms", roomID)

	Client.HSet(c, roomID, map[string]interface{}{
		"name":      reply.RoomName,
		"createdAt": time.Now().Unix(),
	})

	c.JSON(200, gin.H{"roomID": roomID})
}

type RoomsMetadata struct {
	Name      string
	CreatedAt string
}

func ViewRooms(c *gin.Context) {
	RoomIDs, err := Client.SMembers(c, "rooms").Result()

	var rooms []RoomsMetadata

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms!"})
		return
	}

	for _, id := range RoomIDs {
		room, _ := Client.HGetAll(c, id).Result()

		tempRoom := RoomsMetadata{
			Name:      room["name"],
			CreatedAt: room["createdAt"],
		}

		rooms = append(rooms, tempRoom)
	}

	c.JSON(200, gin.H{"rooms": rooms})
}

func JoinRoom() {

}
