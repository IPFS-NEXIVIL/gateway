package main

import (
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	ipfs := router.Group("/ipfs")
	{
		ipfs.GET("/relay-node", start)
		ipfs.POST("/connect-to-local-ipfs", connect)
	}

	router.Run("localhost:8001")
}
