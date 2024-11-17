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
	"strings"
	"time"
)

// User represents the user model
type User struct {
	Name     string `json:"name" binding:"required"`
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

	authorized := r.Group("/")
	authorized.Use(jwtAuthMiddleware())
	{
		r.GET("/profile", jwtAuthMiddleware(), getProfileHandler)
	}

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
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	var storedPassword string
	error := db.QueryRow(context.Background(),
		"SELECT password FROM users WHERE email = $1 AND password = $2",
		loginRequest.Email, loginRequest.Password).Scan(&storedPassword)
	if error == pgx.ErrNoRows {
		// Email not found
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	} else if error != nil {
		// Database error
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error"})
		return
	}

	// Generate and send JWT if credentials are valid
	token, err := generateToken(loginRequest.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create token"})
		return
	}

	// Save the token in the database
	_, err = db.Exec(context.Background(),
		"UPDATE users SET token = $1 WHERE email = $2",
		token, loginRequest.Email,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
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
		"INSERT INTO users (name, email, password) VALUES ($1, $2, $3)",
		user.Name, user.Email, user.Password,
	)

	if err != nil {
		log.Printf("Error inserting user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User signed up successfully!", "name": user.Name})
}

func getProfileHandler(c *gin.Context) {
	// Get email from the context (set by middleware)
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Query database for user profile
	var user User
	err := db.QueryRow(context.Background(),
		"SELECT name, email FROM users WHERE email = $1", email).Scan(&user.Name, &user.Email)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
	}
	// Return the user profile
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get token from authorization header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Authorization header"})
			c.Abort()
			return
		}

		fmt.Println("Authorization header:", authHeader)

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid Authorization header"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims and set in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("email", claims["email"])
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}
