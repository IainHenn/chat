package models

type RoomCreationBody struct {
	RoomName string `json:"room_name"`
	Host     string `json:"host_name"`
}

type RoomsMetadata struct {
	Name      string
	CreatedAt string
}

type RoomActionBody struct {
	RoomName string `json:"room_name"`
	Username string `json:"username"`
}

type ChatMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}
