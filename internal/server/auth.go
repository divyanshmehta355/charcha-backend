package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	// Replace "yourusername" with your actual module name from go.mod
	"github.com/divyanshmehta355/charcha-backend/internal/database"
	"github.com/divyanshmehta355/charcha-backend/internal/models"
)

// We can reuse the RegisterRequest struct, but let's define a LoginRequest for clarity
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse defines what we send back to the React Native app
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// HandleLogin processes user authentication and generates a JWT
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the JSON body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Fetch the user from the database
	var user models.User
	query := `SELECT id, username, password_hash, created_at FROM users WHERE username = $1`

	err := database.DB.QueryRow(context.Background(), query, req.Username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash, // We need this to check the password!
		&user.CreatedAt,
	)

	if err != nil {
		// If the user isn't found, we send a generic error for security
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 4. Compare the provided password with the hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// 5. Generate the JWT Token
	// Create a token with the user's ID and an expiration time (e.g., 24 hours)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	// Sign the token with a secret key from our .env file
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in environment")
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 6. Send the token and user data back to the mobile app
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: tokenString,
		User:  user,
	})
}

// RegisterRequest defines the expected JSON payload from the mobile app
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleRegister processes the user sign-up
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	// 1. Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the JSON body from the React Native app
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 4. Insert the user into Neon PostgreSQL and return the generated ID and timestamps
	query := `
		INSERT INTO users (username, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, username, created_at
	`

	var user models.User
	// We use QueryRow because we expect exactly one row to be returned
	err = database.DB.QueryRow(context.Background(), query, req.Username, string(hashedPassword)).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
	)

	if err != nil {
		log.Printf("Database insert error: %v", err)
		// If the username already exists, Postgres will throw an error based on our UNIQUE constraint
		http.Error(w, "Username might already be taken", http.StatusConflict)
		return
	}

	// 5. Send the successful response back as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(user)
}
