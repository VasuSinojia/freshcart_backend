package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Cart struct {
	Id        int       `json:"id"`
	UserId    string    `json:"user_id"`
	ProductId int       `json:"product_id" required:"true"`
	Quantity  int       `json:"quantity" required:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func addToCart(c *gin.Context) {
	userId, exists := c.Get("id")

	fmt.Println("userId", userId)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var cart Cart
	err := c.BindJSON(&cart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = db.Exec(context.Background(), "INSERT INTO cart (user_id, product_id, quantity) VALUES ($1,$2, $3)", userId, cart.ProductId, cart.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func getCart(c *gin.Context) {
	userId, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rows, err := db.Query(context.Background(), "SELECT * FROM cart WHERE user_id = $1", userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer rows.Close()
	var carts []Cart
	for rows.Next() {
		var cart Cart
		err = rows.Scan(&cart.Id, &cart.UserId, &cart.ProductId, &cart.Quantity, &cart.CreatedAt, &cart.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		carts = append(carts, cart)
	}

	c.JSON(http.StatusOK, gin.H{"cart": carts})
}
