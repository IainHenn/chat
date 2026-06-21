package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

type RoomCreationBody struct {
	RoomName string `json:"room_name"`
	Host     string `json:"host_name"`
}

func CreateRoom(c *gin.Context) {
	var body RoomCreationBody

	if err := c.BindJSON(&body); err != nil {
		fmt.Println(body.RoomName)
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create room, invalid body!"})
		return
	}

	var shorthandName string

	if len(body.RoomName) < 10 {
		shorthandName = body.RoomName
	} else {
		shorthandName = body.RoomName[:10]
	}

	roomID := "room:" + shorthandName

	Client.SAdd(c, "rooms", roomID)

	Client.HSet(c, roomID, map[string]interface{}{
		"name":      body.RoomName,
		"createdAt": time.Now().Unix(),
		"host":      body.Host,
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

type RoomActionBody struct {
	RoomName string `json:"room_name"`
	Username string `json:"username"`
}

func JoinRoom(c *gin.Context) {
	var body RoomActionBody

	if err := c.BindJSON(&body); err != nil {
		fmt.Println(body.RoomName)
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to join room, invalid body!"})
		return
	}

	var shorthandName string

	if len(body.RoomName) < 10 {
		shorthandName = body.RoomName
	} else {
		shorthandName = body.RoomName[:10]
	}

	roomID := "room:" + shorthandName

	RoomIDs, err := Client.SMembers(c, "rooms").Result()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room exists!"})
		return
	}

	found := false

	for _, id := range RoomIDs {
		if roomID == id {
			found = true
			break
		}
	}

	if found == false {
		c.JSON(404, gin.H{"error": "Room not found!"})
		return
	}

	Client.SAdd(c, roomID+":members", body.Username)

	c.JSON(201, gin.H{"roomID": roomID, "username": body.Username})
}

func LeaveRoom(c *gin.Context) {
	var body RoomActionBody

	if err := c.BindJSON(&body); err != nil {
		fmt.Println(body.RoomName)
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to leave room, invalid body!"})
		return
	}

	var shorthandName string

	if len(body.RoomName) < 10 {
		shorthandName = body.RoomName
	} else {
		shorthandName = body.RoomName[:10]
	}

	roomID := "room:" + shorthandName

	Client.SRem(c, roomID+":members", body.Username)

	c.JSON(200, gin.H{"roomID": roomID, "username": body.Username})
}

func DeleteRoom(c *gin.Context) {
	var body RoomActionBody

	if err := c.BindJSON(&body); err != nil {
		fmt.Println(body.RoomName)
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete room, invalid body!"})
		return
	}

	var shorthandName string

	if len(body.RoomName) < 10 {
		shorthandName = body.RoomName
	} else {
		shorthandName = body.RoomName[:10]
	}

	roomID := "room:" + shorthandName

	res, err := Client.HGetAll(c, roomID).Result()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user is owner of room!"})
		return
	}

	if res["host"] != body.Username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to delete room, you're not the owner!"})
		return
	}

	Client.Del(c,
		roomID,
		roomID+":members",
	)

	Client.SRem(c, "rooms", roomID)

	c.JSON(200, gin.H{"deleted": roomID})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // loosen this in production
	},
}

type ChatMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func Ws(c *gin.Context) {
	roomID := c.Query("roomID")
	username := c.Query("username")

	members, err := Client.SMembers(c, roomID+":members").Result()

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to verify if room membership."})
		return
	}

	user_in_room := slices.Contains(members, username)

	if user_in_room == false {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read messages, must join room first!"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println("Failed to connect server via websockets!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect server via websockets!"})
		return
	}

	sub := Client.Subscribe(c, roomID)

	go func() {
		for msg := range sub.Channel() {
			conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			break
		}

		chatMsg, _ := json.Marshal(ChatMessage{
			Username: username,
			Content:  string(msg),
		})

		Client.Publish(c, roomID, chatMsg)
	}

	defer func() {
		sub.Close()
		Client.SRem(c, roomID+":members", username)
		conn.Close()
	}()
}
