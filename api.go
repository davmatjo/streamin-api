package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

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
