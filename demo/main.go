package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	containerId := os.Getenv("NODE_ID")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Container %s] Received a request...", containerId)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "[Container %s] Hello!\n", containerId)
	})
	log.Printf("[Container %s] Starting the server on port :8080", containerId)
	http.ListenAndServe(":8080", mux)

}

