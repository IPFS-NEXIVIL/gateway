package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p/p2p/net/swarm"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
	ma "github.com/multiformats/go-multiaddr"
)

func start(c *gin.Context) {

	listenPort := flag.Int("l", 53100, "wait for incoming connections")
	flag.Parse()

	// Tell the host to relay connections for other peers
	h2, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *listenPort)),
		libp2p.DisableRelay())
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
		relayHost := fmt.Sprintf("%s/p2p/%s\n", ips, h2.ID())
		nodeA(relayHost)
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

func nodeA(relayHost string) {

	relayHost = strings.TrimSpace(relayHost)

	relayAddrInfo, err := addr2info(relayHost)
	if err != nil {
		panic(err)
	}

	// Zero out the listen addresses for the host, so it can only communicate
	// via p2p-circuit
	h3, err := libp2p.New(libp2p.ListenAddrs(), libp2p.EnableRelay())
	if err != nil {
		panic(err)
	}

	if err := h3.Connect(context.Background(), *relayAddrInfo); err != nil {
		panic(err)
	}

	// Now, to test things, let's set up a protocol handler on h3
	h3.SetStreamHandler("/cats", func(s network.Stream) {
		fmt.Println("Meow! It worked!")
		s.Close()
	})

	nodeAID := fmt.Sprintf("%v", h3.ID())

	fmt.Println("Node A ID: ", nodeAID)

	nodeB(relayHost, nodeAID)

	select {}
}

func nodeB(relayHost string, dialID string) {

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

	_, err = h1.NewStream(context.Background(), dialNodeID, "/cats")
	if err == nil {
		fmt.Println("Didnt actually expect to get a stream here. What happened?")
		return
	}
	fmt.Println("Okay, no connection from h1 to h3: ", err)
	fmt.Println("Just as we suspected")

	h1.Network().(*swarm.Swarm).Backoff().Clear(dialNodeID)

	// Creates a relay address to dial ID using relay host as the relay
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

	// Woohoo! we're connected!
	s, err := h1.NewStream(context.Background(), dialNodeID, "/cats")
	if err != nil {
		fmt.Println("huh, this should have worked: ", err)
		return
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream

	fmt.Println("end")

}
