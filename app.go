// This contains the business logic of the backend. There is generally very little here as our backend server is
// relatively dumb, for the most part just forwarding messages to all clients connected via websocket

package main

import (
	"log"
)

type App struct {
	Leader *Client
	Hub    *Hub
	Media  string
}

// NewApp creates a new App and registers it with the given Hub
func NewApp(h *Hub) {
	a := App{
		Hub: h,
	}
	h.Register(a.Receive)
}

// Receive is called whenever the Hub receieves a message. We route the message based on Message MessageType
func (a *App) Receive(m Message) {
	log.Printf("Message: %v", m)

	switch m.Type {
	case Control:
		if m.Subject != a.Leader {
			log.Println("Dumped non-leader control")
		}
		a.Control(m)
	case Info:
		a.Info(m)
	case UserMessage:
		a.UserMessage(m)
	}
}

// Control messages [seek, play, pause] are broadcast to all clients
func (a *App) Control(m Message) {
	a.Hub.Send(Message{
		Type:    Control,
		Subject: nil,
		Action:  m.Action,
		Data:    m.Data,
	})
}

// User Messages are sent to all users (but given a username if they don't have one)
func (a *App) UserMessage(m Message) {
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
func (a *App) Info(m Message) {
	switch m.Action {
	case "register":
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
		return
	case "deregister":
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
	case "users":
		a.SendUsers(m.Subject)
	case "name":
		// Sets the name for a specific client
		log.Printf("Setting name: %s", m.Data)
		m.Subject.Name = m.Data.(string)
		a.SendUsers(nil)
	case "media":
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
		}
	}
}

// SendUsers is mot used by the frontend currently but will send all usernames to a given requester
func (a *App) SendUsers(c *Client) {
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
