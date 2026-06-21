package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/IainHenn/chat/models"
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

func CreateRoom(c *gin.Context) {
	var body models.RoomCreationBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create room, invalid body!"})
		return
	}

	roomID := roomIDFromName(body.RoomName)

	exists, err := roomExists(c, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room!"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Room already exists!"})
		return
	}

	Client.SAdd(c, roomsKey, roomID)
	Client.HSet(c, roomID, map[string]interface{}{
		"name":      body.RoomName,
		"createdAt": time.Now().Unix(),
		"host":      body.Host,
	})
	Client.SAdd(c, membersKey(roomID), body.Host)

	c.JSON(http.StatusOK, gin.H{"roomID": roomID})
}

func ViewRooms(c *gin.Context) {
	roomIDs, err := Client.SMembers(c, roomsKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms!"})
		return
	}

	var rooms []models.RoomsMetadata

	for _, id := range roomIDs {
		if pruneOrphanSetEntry(c, id) {
			continue
		}

		room, err := Client.HGetAll(c, id).Result()
		if err != nil {
			continue
		}

		rooms = append(rooms, models.RoomsMetadata{
			Name:      room["name"],
			CreatedAt: room["createdAt"],
		})
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

func JoinRoom(c *gin.Context) {
	var body models.RoomActionBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to join room, invalid body!"})
		return
	}

	roomID := roomIDFromName(body.RoomName)

	exists, err := roomExists(c, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room exists!"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found!"})
		return
	}

	Client.SAdd(c, membersKey(roomID), body.Username)

	c.JSON(http.StatusCreated, gin.H{"roomID": roomID, "username": body.Username})
}

func LeaveRoom(c *gin.Context) {
	var body models.RoomActionBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to leave room, invalid body!"})
		return
	}

	roomID := roomIDFromName(body.RoomName)

	exists, err := roomExists(c, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room exists!"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found!"})
		return
	}

	member, err := isMember(c, roomID, body.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership!"})
		return
	}
	if !member {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a member of this room!"})
		return
	}

	Client.SRem(c, membersKey(roomID), body.Username)
	cleanupOrphanRoom(c, roomID)

	c.JSON(http.StatusOK, gin.H{"roomID": roomID, "username": body.Username})
}

func DeleteRoom(c *gin.Context) {
	var body models.RoomActionBody

	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete room, invalid body!"})
		return
	}

	roomID := roomIDFromName(body.RoomName)

	res, err := Client.HGetAll(c, roomID).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user is owner of room!"})
		return
	}
	if res["name"] == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found!"})
		return
	}
	if res["host"] != body.Username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to delete room, you're not the owner!"})
		return
	}

	if err := deleteRoom(c, roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": roomID})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // loosen this in production
	},
}

func Ws(c *gin.Context) {
	roomName := c.Query("room_name")
	username := c.Query("username")

	if roomName == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_name and username query params are required!"})
		return
	}

	roomID := roomIDFromName(roomName)

	exists, err := roomExists(c, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room exists!"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found!"})
		return
	}

	member, err := isMember(c, roomID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership!"})
		return
	}
	if !member {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must join room before connecting!"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Failed to connect server via websockets!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect server via websockets!"})
		return
	}

	sub := Client.Subscribe(c, roomID)

	defer func() {
		sub.Close()
		Client.SRem(c, membersKey(roomID), username)
		conn.Close()
	}()

	go func() {
		for msg := range sub.Channel() {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				break
			}
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		chatMsg, err := json.Marshal(models.ChatMessage{
			Username: username,
			Content:  string(msg),
		})
		if err != nil {
			continue
		}

		if err := Client.Publish(c, roomID, chatMsg).Err(); err != nil {
			break
		}
	}
}
