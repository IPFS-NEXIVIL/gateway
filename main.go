package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.tmpl")

	router.POST("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Nexivil File Sharing",
		})
	})
	router.POST("/start", startPubSub)

	router.Run("localhost:8001")
}
