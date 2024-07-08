package main

import (
	"image-server/db"
	"image-server/route"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client := db.ConnectDB()
	db := client.Database(os.Getenv("DB_NAME"))
	port := os.Getenv("PORT")
	r := gin.Default()
	route.Route(r, db)
	r.Run(":" + port)
}
