package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Category struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	ImageUrl string `json:"imageUrl"`
	Color    string `json:"color_code"`
}

type Product struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	CategoryId  int64  `json:"category_id"`
	ImageUrl    string `json:"imageUrl"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

func getCategories(c *gin.Context) {
	var categories []Category

	rows, err := db.Query(context.Background(), "SELECT * FROM categories ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Id, &category.Name, &category.ImageUrl, &category.Color)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while writing data to categories"})
			return
		}
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func getProductsByCategoryId(c *gin.Context) {
	var products []Product
	id := c.Query("category_id")

	rows, err := db.Query(context.Background(), "SELECT * FROM products WHERE category_id = $1", id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}

	defer rows.Close()

	for rows.Next() {
		var product Product
		err := rows.Scan(&product.Id, &product.CategoryId, &product.ImageUrl, &product.Price, &product.Description, &product.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while writing data to products"})
			return
		}
		products = append(products, product)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}
