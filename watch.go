// This contains the business logic of the backend. There is generally very little here as our backend server is
// relatively dumb, for the most part just forwarding messages to all clients connected via websocket

package main

import (
	"encoding/json"
	"log"
)

type WatchSession struct {
	Leader    *Client
	Hub       *Hub
	Media     string
	Thumbnail string
}

// NewWatchSession creates a new WatchSession and registers it with the given Hub
func NewWatchSession(h *Hub) *WatchSession {
	a := WatchSession{
		Hub: h,
	}
	h.Register(a.Receive)
	return &a
}

// Receive is called whenever the Hub receieves a message. We route the message based on Message MessageType
func (a *WatchSession) Receive(m Message) {
	log.Printf("Message: %v", m)

	switch m.Type {
	case Control:
		if m.Subject != a.Leader {
			log.Println("Dumped non-leader control")
		}
		a.control(m)
	case Info:
		a.info(m)
	case UserMessage:
		a.userMessage(m)
	}
}

// Control messages [seek, play, pause] are broadcast to all clients
func (a *WatchSession) control(m Message) {
	a.Hub.Send(Message{
		Type:    Control,
		Subject: nil,
		Action:  m.Action,
		Data:    m.Data,
	})
}

// User Messages are sent to all users (but given a username if they don't have one)
func (a *WatchSession) userMessage(m Message) {
	name := m.Subject.Name
	if name == "" {
		name = "Anonymoose"
	}

	a.Hub.Send(Message{
		Type:    UserMessage,
		Subject: nil,
		Action:  name,
		Data:    m.Data,
	})
}

// Info messages signify information for either the front or backend, so require some additional handling
func (a *WatchSession) info(m Message) {
	switch m.Action {
	case "register":
		a.register(m)
	case "deregister":
		a.deregister(m)
	case "users":
		a.SendUsers(m.Subject)
	case "name":
		a.name(m)
	case "media":
		a.media(m)
	}
}

func (a *WatchSession) media(m Message) {
	if m.Subject == a.Leader {
		a.Media = m.Data.(string)
		a.Hub.Send(Message{
			Type:    "c",
			Subject: nil,
			Action:  "media",
			Data:    m.Data,
		})
		return
	}

	if a.Media != "" {
		a.Hub.Send(Message{
			Type:    "c",
			Subject: m.Subject,
			Action:  "media",
			Data:    a.Media,
		})
		return
	}
}

func (a *WatchSession) name(m Message) {
	// Sets the name for a specific client
	log.Printf("Setting name: %s", m.Data)
	m.Subject.Name = m.Data.(string)
	a.SendUsers(nil)
}

func (a *WatchSession) deregister(m Message) {
	// if the user was the leader we need to elect a new one
	if m.Subject == a.Leader {
		cs := a.Hub.AllClients()
		if len(cs) > 0 {
			a.Leader = a.Hub.AllClients()[0]
			a.Hub.Send(Message{
				Type:    Info,
				Subject: a.Leader,
				Action:  "leader",
				Data:    nil,
			})
		} else {
			a.Leader = nil
		}
	}
	a.Hub.Send(Message{
		Type:    ViewCount,
		Subject: nil,
		Data:    len(a.Hub.AllClients()),
	})
}

func (a *WatchSession) register(m Message) {
	// select a leader if we have none
	if a.Leader == nil {
		a.Leader = m.Subject
		a.Hub.Send(Message{
			Type:    Info,
			Subject: m.Subject,
			Action:  "leader",
			Data:    nil,
		})
	}
	// update all users on the number of viewers
	a.Hub.Send(Message{
		Type:    ViewCount,
		Subject: nil,
		Data:    len(a.Hub.AllClients()),
	})
}

// SendUsers is mot used by the frontend currently but will send all usernames to a given requester
func (a *WatchSession) SendUsers(c *Client) {
	cs := a.Hub.AllClients()
	names := make([]string, 0, len(cs))
	for _, client := range cs {
		if client.Name != "" {
			names = append(names, client.Name)
		}
	}

	a.Hub.Send(Message{
		Type:    "info",
		Subject: c,
		Action:  "users",
		Data:    names,
	})
}

func (a *WatchSession) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id        string
		Media     string
		Thumbnail string
		Viewers   int
	}{
		Id:        a.Hub.id,
		Media:     a.Media,
		Thumbnail: a.Thumbnail,
		Viewers:   len(a.Hub.AllClients()),
	})
}
