package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/divyanshmehta355/charcha-backend/internal/database"
	"github.com/divyanshmehta355/charcha-backend/internal/models"
)

type CreateRoomRequest struct {
	Name    string `json:"name"`
	IsGroup bool   `json:"is_group"`
}

func GetRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `SELECT id, name, is_group, created_at FROM rooms ORDER BY created_at DESC`

	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		log.Printf("Error fetching rooms: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rooms []models.Room

	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.Name, &room.IsGroup, &room.CreatedAt)
		if err != nil {
			log.Printf("Error scanning room row: %v", err)
			continue
		}
		rooms = append(rooms, room)
	}

	if rooms == nil {
		rooms = []models.Room{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO rooms (name, is_group) 
		VALUES ($1, $2) 
		RETURNING id, name, is_group, created_at
	`

	var room models.Room
	err := database.DB.QueryRow(context.Background(), query, req.Name, req.IsGroup).Scan(
		&room.ID,
		&room.Name,
		&room.IsGroup,
		&room.CreatedAt,
	)

	if err != nil {
		log.Printf("Error creating room: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}
