package main

import (
	"log"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create content storage directories
	if err := os.MkdirAll(path.Join("files", "text"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(path.Join("files", "images"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(path.Join("files", "other"), 0755); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	ipfs := router.Group("/ipfs")
	{
		ipfs.POST("/paste", paste)
		ipfs.POST("/upload", upload)
	}

	router.Run("localhost:8001")
}
