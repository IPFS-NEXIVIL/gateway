package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	libp2p "github.com/libp2p/go-libp2p"
)

const Proto = "/relay/echo"

func main() {

	router := gin.Default()
	router.GET("/ipfs", startBrowserNode)

	router.Run("localhost:8001")
}

func startBrowserNode(c *gin.Context) {
	port := flag.Int("l", 9001, "Relay TCP listen port")
	wsport := flag.Int("ws", 9002, "Relay WS listen port")

	// Tell the host use relays
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port),
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", *wsport),
		),
		libp2p.EnableRelay(),
		libp2p.Ping(false),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Relay addresses:\n")
	for _, addr := range host.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", addr.String(), host.ID().Pretty())
	}

	select {}
}
