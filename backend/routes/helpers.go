package routes

import "context"

const roomsKey = "rooms"

// roomIDFromName derives a stable Redis room key from the display name.
// All routes must use this so truncation stays consistent.
func roomIDFromName(roomName string) string {
	if len(roomName) <= 10 {
		return "room:" + roomName
	}
	return "room:" + roomName[:10]
}

func membersKey(roomID string) string {
	return roomID + ":members"
}

func roomExists(ctx context.Context, roomID string) (bool, error) {
	return Client.SIsMember(ctx, roomsKey, roomID).Result()
}

func isMember(ctx context.Context, roomID, username string) (bool, error) {
	return Client.SIsMember(ctx, membersKey(roomID), username).Result()
}

func deleteRoom(ctx context.Context, roomID string) error {
	if err := Client.Del(ctx, roomID, membersKey(roomID)).Err(); err != nil {
		return err
	}
	return Client.SRem(ctx, roomsKey, roomID).Err()
}

func cleanupOrphanRoom(ctx context.Context, roomID string) {
	exists, err := roomExists(ctx, roomID)
	if err != nil || !exists {
		return
	}
	members, err := Client.SMembers(ctx, membersKey(roomID)).Result()
	if err != nil || len(members) > 0 {
		return
	}
	_ = deleteRoom(ctx, roomID)
}

func pruneOrphanSetEntry(ctx context.Context, roomID string) bool {
	room, err := Client.HGetAll(ctx, roomID).Result()
	if err != nil {
		return false
	}
	if room["name"] != "" {
		return false
	}
	_ = deleteRoom(ctx, roomID)
	return true
}
