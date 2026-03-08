package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the Charcha app
type User struct {
	ID       uuid.UUID `json:"id" db:"id"`
	Username string    `json:"username" db:"username"`
	// The "-" ensures the password hash is NEVER accidentally sent to the mobile app
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Room represents a chat conversation (1-on-1 or group)
type Room struct {
	ID uuid.UUID `json:"id" db:"id"`
	// We use a pointer (*string) because Name can be NULL in the DB for 1-on-1 chats
	Name      *string   `json:"name,omitempty" db:"name"`
	IsGroup   bool      `json:"is_group" db:"is_group"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RoomMember represents a user participating in a specific room
type RoomMember struct {
	RoomID   uuid.UUID `json:"room_id" db:"room_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// Message represents a single text message sent in a room
type Message struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RoomID    uuid.UUID `json:"room_id" db:"room_id"`
	SenderID  uuid.UUID `json:"sender_id" db:"sender_id"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
