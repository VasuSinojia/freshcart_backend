package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
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

	var existingQuantity int
	quantityError := db.QueryRow(context.Background(), "SELECT quantity FROM cart WHERE user_id = $1 AND product_id = $2", userId, cart.ProductId).Scan(&existingQuantity)

	if quantityError == nil {
		// Product already exist, update quantity
		_, err := db.Exec(context.Background(), "UPDATE cart SET quantity = quantity + 1 WHERE user_id = $1 AND product_id = $2", userId, cart.ProductId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, err = db.Exec(context.Background(), "INSERT INTO cart (user_id, product_id, quantity) VALUES ($1,$2, 1)", userId, cart.ProductId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quantity"})
			return
		}
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

func removeFromCart(c *gin.Context) {
	userId, exists := c.Get("id")

	productId := c.Query("product_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	_, err := db.Exec(context.Background(), "DELETE FROM cart WHERE user_id = $1 AND product_id = $2", userId, productId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func removeFromCartQuery(productId string, userId float64) {
	_, err := db.Exec(context.Background(), "DELETE FROM cart WHERE user_id = $1 AND product_id = $2", userId, productId)
	if err != nil {
		log.Fatal(err)
	}
}

func incrementQuantity(c *gin.Context) {
	userId, exists := c.Get("id")

	productId := c.Query("product_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	_, err := db.Exec(context.Background(), "UPDATE cart SET quantity = quantity + 1 WHERE user_id = $1 AND product_id = $2", userId, productId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func decrementQuantity(c *gin.Context) {
	userId, exists := c.Get("id")

	productId := c.Query("product_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}
	var quantity int
	err := db.QueryRow(context.Background(), "SELECT quantity FROM cart WHERE user_id = $1 AND product_id = $2", userId, productId).Scan(&quantity)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if quantity == 1 {
		removeFromCartQuery(productId, userId.(float64))
	} else {
		_, err := db.Exec(context.Background(), "UPDATE cart SET quantity = quantity - 1 WHERE user_id = $1 AND product_id = $2", userId, productId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
