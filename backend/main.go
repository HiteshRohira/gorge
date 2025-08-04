package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// In-memory storage for demo purposes
var users []User
var nextID = 1

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get frontend URL for CORS
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Route("/users", func(r chi.Router) {
			r.Get("/", getUsersHandler)
			r.Post("/", createUserHandler)
			r.Get("/{id}", getUserHandler)
		})
	})

	// Initialize with some sample data
	initSampleData()

	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("🌐 Frontend URL: %s\n", frontendURL)
	fmt.Printf("📡 API available at: http://localhost:%s/api\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Success: true,
		Message: "Server is healthy",
		Data: map[string]interface{}{
			"timestamp": time.Now(),
			"version":   "1.0.0",
			"status":    "running",
		},
	}
	writeJSON(w, http.StatusOK, response)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	}
	writeJSON(w, http.StatusOK, response)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := Response{
			Success: false,
			Message: "Invalid request body",
		}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}

	// Basic validation
	if req.Name == "" || req.Email == "" {
		response := Response{
			Success: false,
			Message: "Name and email are required",
		}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}

	// Create new user
	user := User{
		ID:        nextID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}
	nextID++

	users = append(users, user)

	response := Response{
		Success: true,
		Message: "User created successfully",
		Data:    user,
	}
	writeJSON(w, http.StatusCreated, response)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response := Response{
			Success: false,
			Message: "Invalid user ID",
		}
		writeJSON(w, http.StatusBadRequest, response)
		return
	}

	// Find user
	for _, user := range users {
		if user.ID == id {
			response := Response{
				Success: true,
				Message: "User found",
				Data:    user,
			}
			writeJSON(w, http.StatusOK, response)
			return
		}
	}

	response := Response{
		Success: false,
		Message: "User not found",
	}
	writeJSON(w, http.StatusNotFound, response)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

func initSampleData() {
	users = []User{
		{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        2,
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			CreatedAt: time.Now().Add(-12 * time.Hour),
		},
	}
	nextID = 3
}
