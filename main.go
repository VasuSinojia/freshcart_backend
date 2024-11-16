package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"time"
)

// User represents the user model
type User struct {
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

var jwtSecret = []byte("jwt_key")

var db *pgxpool.Pool

func main() {
	var err error

	dbUrl := "postgres://vasu:password@localhost:5432/freshcart?sslmode=disable"
	ctx := context.Background()

	db, err = pgxpool.New(ctx, dbUrl) // Assign to the global variable
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	fmt.Println("Successfully connected to the database!")

	r := gin.Default()
	r.POST("/register", signUpHandler)
	r.POST("/login", loginHandler)
	err = r.Run(":8000")
	if err != nil {
		return
	}
}

func generateToken(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func loginHandler(c *gin.Context) {
	var loginRequest LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var storedPassword string
	error := db.QueryRow(context.Background(),
		"SELECT password FROM users WHERE email = $1 AND password = $2",
		loginRequest.Email, loginRequest.Password).Scan(&storedPassword)
	if error == pgx.ErrNoRows {
		// Email not found
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	} else if error != nil {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate and send JWT if credentials are valid
	token, err := generateToken(loginRequest.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func signUpHandler(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(context.Background(),
		"INSERT INTO users (name, phone, email, password) VALUES ($1, $2, $3, $4)",
		user.Name, user.Phone, user.Email, user.Password,
	)

	if err != nil {
		log.Printf("Error inserting user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User signed up successfully!", "name": user.Name})
}
