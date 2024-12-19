package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Cart struct {
	Id        int    `json:"id"`
	UserId    string `json:"user_id"`
	ProductId int    `json:"product_id" required:"true"`
	Quantity  int    `json:"quantity" required:"true"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
