package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/IainHenn/chat/client/internal/api"
	"github.com/gorilla/websocket"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func RunRoom(serverURL, roomName, username string, client *api.Client) error {
	if err := client.JoinRoom(roomName, username); err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(api.WSURL(serverURL, roomName, username), nil)
	if err != nil {
		_ = client.LeaveRoom(roomName, username)
		return fmt.Errorf("websocket connect: %w", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var msg Message
			if json.Unmarshal(data, &msg) == nil && msg.Username != "" {
				if msg.Username != username {
					fmt.Printf("%s: %s\n", msg.Username, msg.Content)
					os.Stdout.Sync()
				}
			} else {
				fmt.Printf("%s\n", string(data))
				os.Stdout.Sync()
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Connected. Type a message and press Enter. Type /leave to exit.")

	input := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input <- scanner.Text()
		}
		close(input)
	}()

	leave := func() {
		_ = client.LeaveRoom(roomName, username)
		conn.Close()
		<-done
		fmt.Println("Left room.")
	}

	for {
		select {
		case <-sig:
			leave()
			return nil
		case line, ok := <-input:
			if !ok {
				leave()
				return nil
			}
			if strings.TrimSpace(line) == "/leave" {
				leave()
				return nil
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				leave()
				return fmt.Errorf("send message: %w", err)
			}
			fmt.Printf("> %s\n", line)
			os.Stdout.Sync()
		case <-done:
			_ = client.LeaveRoom(roomName, username)
			return fmt.Errorf("disconnected from room")
		}
	}
}
