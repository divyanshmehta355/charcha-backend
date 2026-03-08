package server

import (
	"context"
	"log"
	"sync"

	"github.com/divyanshmehta355/charcha-backend/internal/cache"
)

type Hub struct {
	sync.RWMutex

	Clients map[string]*Client
}

var GlobalHub = &Hub{
	Clients: make(map[string]*Client),
}

func (h *Hub) AddClient(client *Client) {
	h.Lock()
	defer h.Unlock()
	h.Clients[client.UserID] = client
	log.Printf("User %s added to Hub. Total connected: %d", client.UserID, len(h.Clients))
}

func (h *Hub) RemoveClient(userID string) {
	h.Lock()
	defer h.Unlock()
	delete(h.Clients, userID)
	log.Printf("User %s removed from Hub.", userID)
}

func ListenToRedis() {

	pubsub := cache.Client.Subscribe(context.Background(), "charcha_global")
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Println("Background worker is now listening to Upstash Redis...")

	for msg := range ch {

		log.Printf("Redis broadcast received: %s", msg.Payload)

		GlobalHub.RLock()
		for _, client := range GlobalHub.Clients {
			err := client.Conn.WriteMessage(1, []byte(msg.Payload))

			if err != nil {
				log.Printf("Error sending message to %s: %v", client.UserID, err)
			}
		}
		GlobalHub.RUnlock()
	}
}
