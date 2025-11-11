package websocket

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, check origin properly
		return true
	},
}

// ServeWS handles WebSocket requests from clients
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Get user ID from context (if authenticated)
	userID := ""
	if user := r.Context().Value("user_id"); user != nil {
		userID = user.(string)
	}

	// Create new client
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		id:       uuid.New().String(),
		userID:   userID,
		rooms:    make(map[string]bool),
		metadata: make(map[string]interface{}),
	}

	// Register client
	hub.register <- client

	// Send welcome message
	welcomeMsg := Message{
		Type: "welcome",
		Payload: map[string]interface{}{
			"client_id": client.id,
			"message":   "Connected to Marimo ERP WebSocket",
		},
	}
	client.Send(welcomeMsg)

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

// RegisterDefaultHandlers registers default message handlers
func RegisterDefaultHandlers(hub *Hub) {
	// Ping handler
	hub.RegisterHandler("ping", func(client *Client, msg Message) error {
		return client.Send(Message{
			Type: "pong",
			Payload: map[string]interface{}{
				"timestamp": time.Now().Unix(),
			},
		})
	})

	// Join room handler
	hub.RegisterHandler("join", func(client *Client, msg Message) error {
		room, ok := msg.Payload["room"].(string)
		if !ok {
			return fmt.Errorf("invalid room parameter")
		}

		hub.JoinRoom(client, room)

		return client.Send(Message{
			Type: "joined",
			Payload: map[string]interface{}{
				"room": room,
			},
		})
	})

	// Leave room handler
	hub.RegisterHandler("leave", func(client *Client, msg Message) error {
		room, ok := msg.Payload["room"].(string)
		if !ok {
			return fmt.Errorf("invalid room parameter")
		}

		hub.LeaveRoom(client, room)

		return client.Send(Message{
			Type: "left",
			Payload: map[string]interface{}{
				"room": room,
			},
		})
	})

	// Subscribe handler (alias for join)
	hub.RegisterHandler("subscribe", func(client *Client, msg Message) error {
		room, ok := msg.Payload["channel"].(string)
		if !ok {
			return fmt.Errorf("invalid channel parameter")
		}

		hub.JoinRoom(client, room)

		return client.Send(Message{
			Type: "subscribed",
			Payload: map[string]interface{}{
				"channel": room,
			},
		})
	})

	// Unsubscribe handler (alias for leave)
	hub.RegisterHandler("unsubscribe", func(client *Client, msg Message) error {
		room, ok := msg.Payload["channel"].(string)
		if !ok {
			return fmt.Errorf("invalid channel parameter")
		}

		hub.LeaveRoom(client, room)

		return client.Send(Message{
			Type: "unsubscribed",
			Payload: map[string]interface{}{
				"channel": room,
			},
		})
	})
}
