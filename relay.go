package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	network "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	direct "github.com/libp2p/go-libp2p-webrtc-direct"
	mplex "github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	swarm "github.com/libp2p/go-libp2p/p2p/net/swarm"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"
	"github.com/libp2p/go-tcp-transport"
	ma "github.com/multiformats/go-multiaddr"
	webrtc "github.com/pion/webrtc/v3"
)

const protocolID = "/libp2p/circuit/relay/0.1.0"

func start(c *gin.Context) {

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	insecure := flag.Bool("insecure", false, "use an unencrypted connection")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	// Make a host that listens on the given multiaddress
	ha, err := makeRelayHost(*listenF, *insecure, *seed)
	if err != nil {
		log.Fatal(err)
	}

	// Set a stream handler on host A. /echo/1.0.0 is
	// a user-defined protocol name.
	ha.SetStreamHandler(protocolID, func(s network.Stream) {
		log.Println("Got a new stream!")
		if err := doEcho(s); err != nil {
			log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	select {}
}

func makeRelayHost(listenPort int, insecure bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it at least
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(direct.NewTransport(webrtc.Configuration{},
			new(mplex.Transport))),
	)

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/http/p2p-webrtc-direct", listenPort)),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
		transports,
	}

	if insecure {
		opts = append(opts, libp2p.NoSecurity)
	}

	relayHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	_, err = relayv1.NewRelay(relayHost)
	if err != nil {
		log.Fatalf("Failed to instantiate relay: %v", err)
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", relayHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := relayHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if insecure {
		log.Printf("Now run \"./libp2p-echo -l %d -d %s -insecure\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"./libp2p-echo -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return relayHost, nil
}

func addr2info(addrStr string) (*peer.AddrInfo, error) {

	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		panic(err)
	}

	return peer.AddrInfoFromP2pAddr(addr)
}

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

	s, err := h1.NewStream(context.Background(), dialNodeID, protocolID)
	if err != nil {
		fmt.Println("Didnt actually expect to get a stream here. What happened?")
		log.Fatalln(err)
	}

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

	_, err = s.Write([]byte("Hello, world!\n"))
	if err != nil {
		log.Fatalln(err)
	}

	out, err := ioutil.ReadAll(s)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("read reply: %q\n", out)

}

// doEcho reads a line of data a stream and writes it back
func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s\n", str)
	_, err = s.Write([]byte(str))
	return err
}
