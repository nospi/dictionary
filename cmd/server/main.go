package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nospi/dictionary/internal/dictionary"
	"github.com/nospi/dictionary/internal/room"
	"github.com/nospi/dictionary/internal/ws"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize dictionary service (mock for now)
	dict := dictionary.NewMockDictionary()

	// Initialize room manager
	roomManager := room.NewManager()

	// Initialize WebSocket hub with composition
	wsHub := ws.NewHub(roomManager, dict)
	go wsHub.Run()

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(wsHub, w, r)
	})

	// Templates/pages (TODO: implement handlers)
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/room/", handleRoom)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	// TODO: Serve home page template
	w.Write([]byte("Dictionary Game - Coming Soon"))
}

func handleRoom(w http.ResponseWriter, r *http.Request) {
	// TODO: Serve room page template
	w.Write([]byte("Room Page - Coming Soon"))
}
