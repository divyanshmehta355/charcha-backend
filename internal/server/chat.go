package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/divyanshmehta355/charcha-backend/internal/database"
	"github.com/divyanshmehta355/charcha-backend/internal/models"
)

func GetRoomMessages(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		http.Error(w, "Missing room_id parameter", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, room_id, sender_id, content, created_at 
		FROM messages 
		WHERE room_id = $1 
		ORDER BY created_at DESC 
		LIMIT 50
	`

	rows, err := database.DB.Query(context.Background(), query, roomID)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Message

	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning message row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	if messages == nil {
		messages = []models.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
