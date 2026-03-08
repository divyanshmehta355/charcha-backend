package server

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

func ServeWS(w http.ResponseWriter, r *http.Request) {

	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	userID, err := validateToken(tokenString)
	if err != nil {
		log.Printf("Invalid token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		Conn:   ws,
		UserID: userID,
	}

	log.Printf("User %s successfully connected via WebSocket!", client.UserID)

	go client.readMessages()
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

func (c *Client) readMessages() {

	defer func() {
		log.Printf("User %s disconnected", c.UserID)
		c.Conn.Close()
	}()

	for {

		messageType, payload, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", c.UserID, err)
			break

		}

		log.Printf("Received message from %s: %s", c.UserID, string(payload))

		err = c.Conn.WriteMessage(messageType, payload)
		if err != nil {
			log.Printf("Error writing message to %s: %v", c.UserID, err)
			break
		}
	}
}
