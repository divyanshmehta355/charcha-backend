package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/divyanshmehta355/charcha-backend/internal/cache"
	"github.com/divyanshmehta355/charcha-backend/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true

	},
}

type Client struct {
	Conn   *websocket.Conn
	UserID string
}

type IncomingMessage struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

func validateToken(tokenString string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	userID := claims["user_id"].(string)
	return userID, nil
}

func ServeWS(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	userID, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}

	client := &Client{
		Conn:   ws,
		UserID: userID,
	}

	GlobalHub.AddClient(client)

	go client.readMessages()
}

func (c *Client) readMessages() {
	defer func() {
		GlobalHub.RemoveClient(c.UserID)
		c.Conn.Close()
	}()

	for {

		_, payload, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("User %s disconnected or error: %v", c.UserID, err)
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(payload, &incoming); err != nil {
			log.Printf("Invalid message format from %s: %v", c.UserID, err)
			continue

		}

		query := `
			INSERT INTO messages (room_id, sender_id, content) 
			VALUES ($1, $2, $3) 
			RETURNING id::text, created_at
		`

		var messageID string
		var createdAt time.Time

		err = database.DB.QueryRow(context.Background(), query, incoming.RoomID, c.UserID, incoming.Content).Scan(&messageID, &createdAt)
		if err != nil {
			log.Printf("Failed to save message to DB: %v", err)
			continue
		}

		broadcastData := map[string]string{
			"id":        messageID,
			"room_id":   incoming.RoomID,
			"sender_id": c.UserID,
			"content":   incoming.Content,

			"created_at": createdAt.Format(time.RFC3339),
		}

		jsonBroadcast, _ := json.Marshal(broadcastData)

		err = cache.Client.Publish(context.Background(), "charcha_global", jsonBroadcast).Err()
		if err != nil {
			log.Printf("Error publishing to Redis: %v", err)
		} else {
			log.Printf("Message %s saved to DB and published to Redis!", messageID)
		}
	}
}
