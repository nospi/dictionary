package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nospi/dictionary/internal/api"
	"github.com/nospi/dictionary/internal/dictionary"
	"github.com/nospi/dictionary/internal/room"
	"github.com/nospi/dictionary/internal/sse"
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

	// Initialize SSE hub
	sseHub := sse.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sseHub.Run(ctx)

	// Load templates
	tmpl, err := loadTemplates()
	if err != nil {
		log.Fatal("Failed to load templates: ", err)
	}

	// Initialize API server
	apiServer := api.NewServer(roomManager, sseHub, dict, tmpl)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API endpoints
	mux.HandleFunc("/api/room/create", apiServer.CreateRoomHandler)
	mux.HandleFunc("/api/room/join", apiServer.JoinRoomHandler)
	mux.HandleFunc("/api/game/start", apiServer.StartGameHandler)
	mux.HandleFunc("/api/game/definition", apiServer.SubmitDefinitionHandler)
	mux.HandleFunc("/api/game/vote", apiServer.SubmitVoteHandler)
	mux.HandleFunc("/api/game/end", apiServer.EndGameHandler)
	mux.HandleFunc("/api/events", apiServer.SSEHandler)
	mux.HandleFunc("/api/health", apiServer.HealthHandler)

	// Page routes (more specific routes first)
	mux.HandleFunc("/room/", handleRoom(tmpl, roomManager))
	mux.HandleFunc("/", handleHome(tmpl))

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // No timeout for SSE
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		cancel() // Stop SSE hub
	}()

	log.Printf("Server starting on port %s", port)
	log.Printf("Visit http://localhost:%s to play", port)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe: ", err)
	}

	log.Println("Server stopped")
}

// loadTemplates loads all HTML templates with helper functions
func loadTemplates() (*template.Template, error) {
	// Define template helper functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"divf": func(a, b int) float64 {
			return float64(a) / float64(b)
		},
		"mulf": func(a float64, b float64) float64 {
			return a * b
		},
		"maxScore": func(scores map[string]int) int {
			max := 0
			for _, score := range scores {
				if score > max {
					max = score
				}
			}
			return max
		},
	}

	// Get all template files
	pattern := filepath.Join("web", "templates", "*.html")
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	// Also load partials
	partialPattern := filepath.Join("web", "templates", "partials", "*.html")
	tmpl, err = tmpl.ParseGlob(partialPattern)
	if err != nil {
		// Partials might not exist yet, that's ok
		log.Printf("Note: No partials found at %s", partialPattern)
	}

	return tmpl, nil
}

// handleHome serves the home page
func handleHome(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handleHome called: path=%s", r.URL.Path)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "home.html", nil); err != nil {
			log.Printf("Error rendering home template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// handleRoom serves the room page
func handleRoom(tmpl *template.Template, roomMgr *room.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handleRoom called: path=%s", r.URL.Path)
		// Extract room code from URL path (/room/ABC123)
		roomCode := r.URL.Path[len("/room/"):]
		if roomCode == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Get room
		rm, err := roomMgr.GetRoom(roomCode)
		if err != nil {
			// Room not found
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			if err := tmpl.ExecuteTemplate(w, "error.html", map[string]string{
				"Title":   "Room Not Found",
				"Message": "The room code '" + roomCode + "' does not exist. It may have been closed.",
			}); err != nil {
				log.Printf("Error rendering error template: %v", err)
				http.Error(w, "Room not found", http.StatusNotFound)
			}
			return
		}

		// Get player ID from cookie
		var playerID string
		if cookie, err := r.Cookie("player_id"); err == nil {
			playerID = cookie.Value
		}

		// Prepare template data
		data := map[string]interface{}{
			"Room":     rm,
			"PlayerID": playerID,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "room.html", data); err != nil {
			log.Printf("Error rendering room template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
