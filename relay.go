package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	direct "github.com/libp2p/go-libp2p-webrtc-direct"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	swarm "github.com/libp2p/go-libp2p/p2p/net/swarm"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pion/webrtc/v3"
)

func start(c *gin.Context) {

	listenPort := flag.Int("l", 53100, "wait for incoming connections")
	flag.Parse()

	// libp2p webrtc
	transport := direct.NewTransport(
		webrtc.Configuration{},
		new(mplex.Transport),
	)

	// Tell the host to relay connections for other peers
	h2, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *listenPort)),
		libp2p.DisableRelay(),
		libp2p.Transport(transport),
	)
	if err != nil {
		log.Printf("Failed to create h2: %v", err)
		return
	}
	_, err = relayv1.NewRelay(h2)
	if err != nil {
		log.Printf("Failed to instantiate h2 relay: %v", err)
		return
	}

	for _, ips := range h2.Addrs() {
		fmt.Printf("%s/p2p/%s\n", ips, h2.ID())
	}

	fmt.Printf("%s/p2p/%s\n", h2.Addrs()[len(h2.Addrs())-1], h2.ID())

	select {}

}

func addr2info(addrStr string) (*peer.AddrInfo, error) {

	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		panic(err)
	}

	return peer.AddrInfoFromP2pAddr(addr)
}

const protocolID = "/libp2p/circuit/relay/0.1.0"

func connect(c *gin.Context) {

	type ContentRequestBody struct {
		RelayHost string `json:"relay_host"`
		DialID    string `json:"dial_id"`
	}

	var requestBody ContentRequestBody

	c.BindJSON(&requestBody)

	var relayHost = requestBody.RelayHost
	var dialID = requestBody.DialID

	relayAddrInfo, err := addr2info(relayHost)
	if err != nil {
		panic(err)
	}

	h1, err := libp2p.New(libp2p.EnableRelay())
	if err != nil {
		panic(err)
	}

	if err := h1.Connect(context.Background(), *relayAddrInfo); err != nil {
		panic(err)
	}

	dialNodeID, err := peer.Decode(dialID)

	if err != nil {
		panic(err)
	}

	_, err = h1.NewStream(context.Background(), dialNodeID, protocolID)
	if err == nil {
		fmt.Println("Didnt actually expect to get a stream here. What happened?")
		return
	}
	fmt.Println("Okay, no connection from h1 to h3: ", err)
	fmt.Println("Just as we suspected")

	h1.Network().(*swarm.Swarm).Backoff().Clear(dialNodeID)

	relayaddr, err := ma.NewMultiaddr(relayHost + "/p2p-circuit/p2p/" + dialNodeID.Pretty())

	if err != nil {
		panic(err)
	}

	h3relayInfo := peer.AddrInfo{
		ID:    dialNodeID,
		Addrs: []ma.Multiaddr{relayaddr},
	}

	if err := h1.Connect(context.Background(), h3relayInfo); err != nil {
		panic(err)
	}

	fmt.Println("end")

}
