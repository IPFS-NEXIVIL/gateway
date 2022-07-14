package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	shell "github.com/ipfs/go-ipfs-api"
)

func main() {

	router := gin.Default()
	router.GET("/ipfs/:cid", getFromIPFS)
	router.POST("/ipfs", setToIPFS)

	router.Run("localhost:8000")
}

func getFromIPFS(c *gin.Context) {
	cid := c.Param("cid")

	// Where your local node is running on localhost:5002
	sh := shell.NewShell("localhost:5002")

	catResult, err := sh.Cat(cid)
	if err != nil {
		fmt.Println(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(catResult)
	// resp := buf.String()

	c.Data(http.StatusOK, "text/html", buf.Bytes())

}

func setToIPFS(c *gin.Context) {

	type ContentRequestBody struct {
		Content string `json:"content"`
	}

	var requestBody ContentRequestBody

	c.BindJSON(&requestBody)

	cid, err := set(requestBody.Content)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	c.JSON(http.StatusOK, gin.H{"content_cid": cid})

}

func set(data string) (string, error) {
	// Where your local node is running on localhost:5002
	sh := shell.NewShell("localhost:5002")

	cid, err := sh.Add(strings.NewReader(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return "", err
	}

	fmt.Printf("added %s", cid)

	return cid, nil
}
