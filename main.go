package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

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
		r.GET("/categories", jwtAuthMiddleware(), getCategories)
		r.GET("/products", jwtAuthMiddleware(), getProductsByCategoryId)
		r.POST("/cart", jwtAuthMiddleware(), addToCart)
		r.GET("/cart", jwtAuthMiddleware(), getCart)
		r.DELETE("/cart", jwtAuthMiddleware(), removeFromCart)
	}

	err = r.Run(":8000")
	if err != nil {
		return
	}
}
