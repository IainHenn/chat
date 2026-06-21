package routes

import "testing"

func TestRoomIDFromName(t *testing.T) {
	tests := []struct {
		name     string
		roomName string
		want     string
	}{
		{"short name", "general", "room:general"},
		{"exactly ten chars", "1234567890", "room:1234567890"},
		{"truncated", "my-long-room-name", "room:my-long-ro"},
		{"curl test case", "curl-test-room", "room:curl-test-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roomIDFromName(tt.roomName); got != tt.want {
				t.Errorf("roomIDFromName(%q) = %q, want %q", tt.roomName, got, tt.want)
			}
		})
	}
}
