package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/divyanshmehta355/charcha-backend/internal/cache"
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
			break
		}

		messageData := map[string]string{
			"sender_id": c.UserID,
			"content":   string(payload),
		}

		jsonPayload, _ := json.Marshal(messageData)

		err = cache.Client.Publish(context.Background(), "charcha_global", jsonPayload).Err()
		if err != nil {
			log.Printf("Error publishing to Redis: %v", err)
		}
	}
}
