package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"log"
	"net/http"
	"os"
	"sort"
)

type App struct {
	Instances map[string]*WatchSession
}

func NewApp() *App {
	return &App{
		Instances: make(map[string]*WatchSession),
	}
}

func (a *App) HandleCheckSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if _, ok := a.Instances[id]; ok {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(404)
	}
}

func (a *App) HandleJoinSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if ws, ok := a.Instances[id]; ok {
		serveWs(ws.Hub, w, r)
	} else {
		w.WriteHeader(404)
	}
}

func (a *App) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	resp := struct {
		Sessions []*WatchSession
	}{}

	sessions := make([]*WatchSession, 0, len(a.Instances))
	for _, s := range a.Instances {
		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Hub.id < sessions[j].Hub.id
	})
	resp.Sessions = sessions

	b, _ := json.Marshal(resp)
	w.WriteHeader(200)
	w.Write(b)
}

func (a *App) HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	id := xid.New().String()

	if _, ok := a.Instances[id]; ok {
		// This should never realistically happen, but we will fail anyway *just* in case
		w.WriteHeader(500)
		return
	}

	h := newHub(id)
	go h.run()
	ws := NewWatchSession(h)
	a.Instances[id] = ws

	w.Header().Add("Location", id)
	w.WriteHeader(201)
}

func (a *App) HandleDestroySession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if inst, ok := a.Instances[id]; ok {
		delete(a.Instances, id)
		inst.Hub.Close()
		w.WriteHeader(200)
		return
	} else {
		w.WriteHeader(404)
		return
	}
}

func ListMedia(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("./media")
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	list, err := file.Readdirnames(0)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	resp := struct {
		Items []string
	}{
		list,
	}

	b, _ := json.Marshal(resp)
	w.WriteHeader(200)
	w.Write(b)
}
