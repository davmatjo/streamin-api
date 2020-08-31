// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

var (
	addr = flag.String("addr", ":8080", "http service address")
)

func main() {
	flag.Parse()

	hub := newHub()
	go hub.run()

	r := mux.NewRouter()

	// Websocket
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	// API
	r.HandleFunc("/api/media", ListMedia).Methods("GET")

	r.HandleFunc("/privacy", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Your google profile and email will be stored when you login with google"))
	}).Methods("GET")

	// File Server
	r.PathPrefix("/media/").Handler(http.StripPrefix("/media/", http.FileServer(http.Dir("./media")))).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web"))).Methods("GET")

	NewApp(hub)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
