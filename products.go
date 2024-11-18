package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Category struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func getCategories(c *gin.Context) {
	var categories []Category

	rows, err := db.Query(context.Background(), "SELECT * FROM categories")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.Id, &category.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error in write data to categories"})
			return
		}
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
