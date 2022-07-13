package main

import "github.com/gin-gonic/gin"

func main() {

	router := gin.Default()

	ipfs := router.Group("/ipfs")
	{
		ipfs.GET("/relay-start", start)
	}

	router.Run("localhost:8001")
}
