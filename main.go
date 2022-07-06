package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p"
)

func main() {
	// start a libp2p node that listens on TCP port 2000 on the IPv4
	// loopback interface
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/2000"),
	)
	if err != nil {
		panic(err)
	}

	// print the node's listening addresses
	fmt.Println("Listen addresses:", node.Addrs())

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
