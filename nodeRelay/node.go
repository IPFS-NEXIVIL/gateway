package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
)

func main() {

	listenPort := flag.Int("l", 53100, "wait for incoming connections")
	flag.Parse()

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
	}

	fmt.Printf("%s/p2p/%s\n", h2.Addrs()[len(h2.Addrs())-1], h2.ID())

	select {}

}
