package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	addr = flag.String("addr", ":8080", "http service address")
)

func main() {
	flag.Parse()

	app := NewApp()

	r := mux.NewRouter()

	// API
	r.HandleFunc("/api/media", ListMedia).Methods("GET")
	r.HandleFunc("/api/sessions", app.HandleListSessions).Methods("GET")
	r.HandleFunc("/api/sessions/create", app.HandleCreateSession).Methods("POST")
	r.HandleFunc("/api/sessions/{id}/join", app.HandleJoinSession)
	r.HandleFunc("/api/sessions/{id}", app.HandleCheckSession).Methods("GET")
	r.HandleFunc("/api/sessions/{id}", app.HandleDestroySession).Methods("DELETE")

	r.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Your google profile and email will be stored when you login with google"))
	}).Methods("GET")

	// File Server
	r.PathPrefix("/media/").Handler(http.StripPrefix("/media/", http.FileServer(http.Dir("./media")))).Methods("GET")
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		path = filepath.Join("./web", path)
		_, err = os.Stat(path)
		if err != nil {
			http.ServeFile(w, r, "./web/index.html")
			return
		}

		http.FileServer(http.Dir("./web")).ServeHTTP(w, r)
	}).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
