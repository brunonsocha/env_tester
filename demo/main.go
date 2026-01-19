package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	containerId := os.Getenv("NODE_ID")
	if containerId == "" {
		var err error
		containerId, err = os.Hostname()
		if err != nil {
			log.Fatalf("[Error] %v", err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Container %s] Received a request...", containerId)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "[Container %s] Hello!\n", containerId)
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /id", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, containerId)
	})
	http.ListenAndServe(":8080", mux)

}


