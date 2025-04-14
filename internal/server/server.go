package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centrifugal/centrifuge"
)

// Server represents our WebSocket server that handles real-time video chat communication
// using Centrifuge as the underlying WebSocket framework. Centrifuge provides:
// - Reliable WebSocket connections
// - Channel-based pub/sub messaging
// - Message history and recovery
// - Scalable architecture for handling multiple concurrent connections
type Server struct {
	node *centrifuge.Node
}

// Message defines the structure of messages exchanged between clients
// Type: Indicates the message type (join-room, offer, answer, ice-candidate)
// RoomID: Identifies the chat room
// UserID: Identifies the sender
// ToUserID: Identifies the recipient for peer-to-peer messages
// Payload: Contains the actual WebRTC signaling data (SDP offers/answers, ICE candidates)
type Message struct {
	Type     string      `json:"type"`
	RoomID   string      `json:"roomId,omitempty"`
	UserID   string      `json:"userId,omitempty"`
	ToUserID string      `json:"toUserId,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
}

func NewServer() *Server {
	// Initialize Centrifuge node with debug logging enabled
	node, err := centrifuge.New(centrifuge.Config{
		LogLevel: centrifuge.LogLevelDebug,
	})
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		node: node,
	}

	// Configure node handlers for client lifecycle events
	node.OnConnect(func(client *centrifuge.Client) {
		log.Printf("Client connected: %s", client.ID())

		// Set up client event handlers for subscription and message publishing
		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			// When a client subscribes to a room channel, we allow the subscription
			// This enables them to receive messages from other participants
			log.Printf("Client %s subscribes to %s", client.ID(), e.Channel)
			cb(centrifuge.SubscribeReply{}, nil)
		})

		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			// Handle incoming WebRTC signaling messages from clients
			var msg Message
			if err := json.Unmarshal(e.Data, &msg); err != nil {
				cb(centrifuge.PublishReply{}, err)
				return
			}

			// Route different types of WebRTC signaling messages to appropriate handlers
			switch msg.Type {
			case "join-room":
				s.handleJoinRoom(client, msg)
			case "offer":
				s.handleOffer(msg)
			case "answer":
				s.handleAnswer(msg)
			case "ice-candidate":
				s.handleICECandidate(msg)
			}

			cb(centrifuge.PublishReply{}, nil)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			log.Printf("Client disconnected: %s", client.ID())
		})
	})

	return s
}

// handleJoinRoom manages room membership and notifies other participants
// When a user joins a room:
// 1. They are subscribed to the room's channel
// 2. Other participants are notified of the new user
// 3. The room maintains message history for new participants
func (s *Server) handleJoinRoom(client *centrifuge.Client, msg Message) {
	channel := "room:" + msg.RoomID

	// Notify room participants about the new user
	joinMsg := Message{
		Type:   "user-connected",
		UserID: msg.UserID,
	}
	data, _ := json.Marshal(joinMsg)

	// Publish with history enabled to allow new participants to see recent messages
	s.node.Publish(channel, data, centrifuge.WithHistory(10, 24*time.Hour))

	// Subscribe the client to the room channel to receive future messages
	client.Subscribe(channel)
}

// handleOffer forwards WebRTC SDP offers between peers
// Uses Centrifuge's pub/sub to deliver the offer to the specific recipient
func (s *Server) handleOffer(msg Message) {
	data, _ := json.Marshal(Message{
		Type:    "offer",
		UserID:  msg.UserID,
		Payload: msg.Payload,
	})
	s.node.Publish("user:"+msg.ToUserID, data)
}

// handleAnswer forwards WebRTC SDP answers between peers
// Uses Centrifuge's pub/sub to deliver the answer to the specific recipient
func (s *Server) handleAnswer(msg Message) {
	data, _ := json.Marshal(Message{
		Type:    "answer",
		UserID:  msg.UserID,
		Payload: msg.Payload,
	})
	s.node.Publish("user:"+msg.ToUserID, data)
}

// handleICECandidate forwards WebRTC ICE candidates between peers
// Uses Centrifuge's pub/sub to deliver the ICE candidate to the specific recipient
func (s *Server) handleICECandidate(msg Message) {
	data, _ := json.Marshal(Message{
		Type:    "ice-candidate",
		UserID:  msg.UserID,
		Payload: msg.Payload,
	})
	s.node.Publish("user:"+msg.ToUserID, data)
}

// Start initializes the WebSocket server and begins listening for connections
// Sets up CORS to allow connections from any origin (for demo purposes)
func (s *Server) Start(port string) error {
	if err := s.node.Run(); err != nil {
		return err
	}

	wsHandler := centrifuge.NewWebsocketHandler(s.node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	})

	http.Handle("/connection/websocket", wsHandler)
	return http.ListenAndServe(port, nil)
}
