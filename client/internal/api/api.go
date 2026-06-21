package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: http.DefaultClient,
	}
}

type apiError struct {
	Error string `json:"error"`
}

type roomListResponse struct {
	Rooms []struct {
		Name      string `json:"Name"`
		CreatedAt string `json:"CreatedAt"`
	} `json:"rooms"`
}

func (c *Client) do(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp apiError
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) CreateRoom(roomName, host string) error {
	return c.do(http.MethodPost, "/rooms", map[string]string{
		"room_name": roomName,
		"host_name": host,
	}, nil)
}

func (c *Client) ViewRooms() ([]struct {
	Name      string
	CreatedAt string
}, error) {
	var resp roomListResponse
	if err := c.do(http.MethodGet, "/rooms", nil, &resp); err != nil {
		return nil, err
	}
	rooms := make([]struct {
		Name      string
		CreatedAt string
	}, len(resp.Rooms))
	for i, r := range resp.Rooms {
		rooms[i].Name = r.Name
		rooms[i].CreatedAt = r.CreatedAt
	}
	return rooms, nil
}

func (c *Client) JoinRoom(roomName, username string) error {
	return c.do(http.MethodPut, "/rooms/join", map[string]string{
		"room_name": roomName,
		"username":  username,
	}, nil)
}

func (c *Client) LeaveRoom(roomName, username string) error {
	return c.do(http.MethodPut, "/rooms/leave", map[string]string{
		"room_name": roomName,
		"username":  username,
	}, nil)
}

func (c *Client) DeleteRoom(roomName, username string) error {
	return c.do(http.MethodDelete, "/rooms", map[string]string{
		"room_name": roomName,
		"username":  username,
	}, nil)
}

func WSURL(httpBase, roomName, username string) string {
	wsBase := strings.Replace(httpBase, "https://", "wss://", 1)
	wsBase = strings.Replace(wsBase, "http://", "ws://", 1)
	q := url.Values{}
	q.Set("room_name", roomName)
	q.Set("username", username)
	return wsBase + "/ws?" + q.Encode()
}
