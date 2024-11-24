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
