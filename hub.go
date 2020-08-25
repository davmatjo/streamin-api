// Parts of this source are Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the client.go file.

// The original source has been modified to pass all messages to any application that registers themselves as
// a listener with the Register method. Additionally, sending logic has been modified to allow both targeted and
// broadcasted messages.
package main

import (
	"encoding/json"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Inbound

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Registered Listeners within this application
	listeners []func(Message)
}

type MessageType string

const (
	Control     MessageType = "c"
	Info        MessageType = "i"
	UserMessage MessageType = "m"
	ViewCount   MessageType = "v"
)

// Message provides the message protocol along with some additional recipient/sender metadata
type Message struct {
	Type MessageType
	// Subject contains the sender for inbound messages. For outbound messages it describes the recipient.
	// Subject can be nil for outbound messages which specified a broadcast message
	Subject *Client `json:"-"`
	// Action provides additional context to the message type
	Action string
	// Data may contain arbitrary structured or unstructured data depending on the Message Type
	Data interface{}
}

func (m Message) Marshal() []byte {
	b, _ := json.Marshal(m)
	return b
}

func ParseMessage(b []byte) (m Message, err error) {
	err = json.Unmarshal(b, &m)
	return m, err
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Inbound),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		// On ws connect or disconnect we create a Message to inform any listening apps of the change
		case client := <-h.register:
			h.clients[client] = true
			sendToListeners(h.listeners, Message{
				Type:    Info,
				Subject: client,
				Action:  "register",
				Data:    "",
			})
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				sendToListeners(h.listeners, Message{
					Type:    Info,
					Subject: client,
					Action:  "deregister",
					Data:    "",
				})
			}
		// All inbound messages are parsed into Messages and sent to all listeners
		case message := <-h.broadcast:
			m, err := ParseMessage(message.Message)
			m.Subject = message.Client
			if err != nil {
				log.Printf("Recieved Invalid Message: %s", string(message.Message))
			}
			sendToListeners(h.listeners, m)
		}
	}
}

func sendToListeners(fs []func(Message), m Message) {
	for _, f := range fs {
		f(m)
	}
}

// Sends the message according to the Client rules described in Message
func (h *Hub) Send(message Message) {
	if message.Subject == nil {
		for client := range h.clients {
			select {
			case client.send <- message.Marshal():
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
		return
	}

	if _, ok := h.clients[message.Subject]; ok {
		message.Subject.send <- message.Marshal()
	}
}

// Register a new app that is interested in receiving incoming messages
func (h *Hub) Register(f func(Message)) {
	h.listeners = append(h.listeners, f)
}

// Get all currently connected websocket clients
func (h *Hub) AllClients() []*Client {
	cs := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		if client != nil {
			cs = append(cs, client)
		}
	}
	return cs
}
