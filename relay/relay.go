package relay

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type relayRequest struct {
	nodeID     string   `json:"node_id"`
	netID      string   `json:"net_id"`
	relayAddrs []string `json:"relay_addrs"`
}

func processRelayCreation(c *gin.Context) {
	var requestBody relayRequest

	c.BindJSON(&requestBody)

	relayNodeID := requestBody.nodeID

}

func createRelay(w http.ResponseWriter, r *http.Request) {
	var relay relayRequest

	w.Header().Set("Content-Type", "application/json")

}
