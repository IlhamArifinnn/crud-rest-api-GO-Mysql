package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
)

var validate = validator.New()
var db *sql.DB

func initDB() {
	var err error
	// Format: username:password@tcp(host:port)/dbname
	db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/db_albums")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connected")
}

type album struct {
	ID    string  `json:"id" validate:"required"`
	Title string  `json:"title" validate:"required"`
	Price float64 `json:"price" validate:"required,gte=0"`
}

var albums = []album{
	{ID: "1", Title: "tes1", Price: 10.99},
	{ID: "2", Title: "tes2", Price: 11.99},
	{ID: "3", Title: "tes3", Price: 12.99},
}

func getAlbums(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, price FROM albums")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var albums []album
	for rows.Next() {
		var alb album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Price); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		albums = append(albums, alb)
	}

	c.JSON(http.StatusOK, albums)
}

func postAlbums(c *gin.Context) {
	var newAlbum album

	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if err := validate.Struct(newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("INSERT INTO albums (id, title, price) VALUES (?, ?, ?)", newAlbum.ID, newAlbum.Title, newAlbum.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newAlbum)
}

func getAlbumById(c *gin.Context) {
	id := c.Param("id")

	var alb album
	err := db.QueryRow("SELECT id, title, price FROM albums WHERE id = ?", id).Scan(&alb.ID, &alb.Title, &alb.Price)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"message": "Album not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alb)
}

func updateAlbum(c *gin.Context) {
	id := c.Param("id")
	var updatedAlbum album

	if err := c.BindJSON(&updatedAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	_, err := db.Exec("UPDATE albums SET title = ?, price = ? WHERE id = ?", updatedAlbum.Title, updatedAlbum.Price, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Album updated"})
}

func deleteAlbum(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM albums WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Album deleted"})
}

func main() {
	initDB() // Initialize database connection
	defer db.Close()

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.POST("/albums", postAlbums)
	router.GET("/albums/:id", getAlbumById)
	router.PUT("/albums/:id", updateAlbum)
	router.DELETE("/albums/:id", deleteAlbum)

	if err := router.Run("localhost:8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
