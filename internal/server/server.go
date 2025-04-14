package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centrifugal/centrifuge"
)

type Server struct {
	node *centrifuge.Node
}

type Message struct {
	Type     string      `json:"type"`
	RoomID   string      `json:"roomId,omitempty"`
	UserID   string      `json:"userId,omitempty"`
	ToUserID string      `json:"toUserId,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
}

func NewServer() *Server {
	node, err := centrifuge.New(centrifuge.Config{
		LogLevel: centrifuge.LogLevelDebug,
	})
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		node: node,
	}

	// Configure node handlers
	node.OnConnect(func(client *centrifuge.Client) {
		log.Printf("Client connected: %s", client.ID())

		// Set up client event handlers
		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			log.Printf("Client %s subscribes to %s", client.ID(), e.Channel)
			cb(centrifuge.SubscribeReply{}, nil)
		})

		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			var msg Message
			if err := json.Unmarshal(e.Data, &msg); err != nil {
				cb(centrifuge.PublishReply{}, err)
				return
			}

			switch msg.Type {
			case "join-room":
				s.handleJoinRoom(client, msg)
			case "offer":
				s.handleOffer(client, msg)
			case "answer":
				s.handleAnswer(client, msg)
			case "ice-candidate":
				s.handleICECandidate(client, msg)
			}

			cb(centrifuge.PublishReply{}, nil)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			log.Printf("Client disconnected: %s", client.ID())
		})
	})

	return s
}

func (s *Server) handleJoinRoom(client *centrifuge.Client, msg Message) {
	channel := "room:" + msg.RoomID

	// Publish user joined event to the room
	joinMsg := Message{
		Type:   "user-connected",
		UserID: msg.UserID,
	}
	data, _ := json.Marshal(joinMsg)

	s.node.Publish(channel, data, centrifuge.WithHistory(10, 24*time.Hour))

	// Subscribe client to the room channel
	client.Subscribe(channel)
}

func (s *Server) handleOffer(client *centrifuge.Client, msg Message) {
	data, _ := json.Marshal(Message{
		Type:    "offer",
		UserID:  msg.UserID,
		Payload: msg.Payload,
	})
	s.node.Publish("user:"+msg.ToUserID, data)
}

func (s *Server) handleAnswer(client *centrifuge.Client, msg Message) {
	data, _ := json.Marshal(Message{
		Type:    "answer",
		UserID:  msg.UserID,
		Payload: msg.Payload,
	})
	s.node.Publish("user:"+msg.ToUserID, data)
}

func (s *Server) handleICECandidate(client *centrifuge.Client, msg Message) {
	data, _ := json.Marshal(Message{
		Type:    "ice-candidate",
		UserID:  msg.UserID,
		Payload: msg.Payload,
	})
	s.node.Publish("user:"+msg.ToUserID, data)
}

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
