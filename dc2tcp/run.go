package dc2tcp

import (
	signaling_client "github.com/peergoim/signaling-client"
	"github.com/peergoim/tcp2p/dc2tcp/internal/config"
	"github.com/peergoim/tcp2p/dc2tcp/internal/handler"
	"github.com/peergoim/tcp2p/dc2tcp/internal/svc"
	"strings"
)

func Run(c *config.Config) {
	if len(c.StunUrls) == 0 {
		c.StunUrls = config.DefaultStunUrls
	}
	if c.TcpAddr == "" {
		panic("tcp addr is empty")
	} else if len(strings.Split(c.TcpAddr, ":")) != 2 {
		panic("tcp addr is invalid")
	}
	if c.PeerId == "" {
		panic("peer id is empty")
	}
	if !strings.HasPrefix(c.SignalingEndpoint, "ws") {
		panic("signaling endpoint must start with ws or wss")
	}
	ctx := svc.NewServiceContext(c)
	handers := map[string]signaling_client.MethodHandler{
		"offer": func(request *signaling_client.CallRequest) *signaling_client.CallResponse {
			offerHandler := handler.NewOfferHandler(ctx, request)
			return offerHandler.Handle()
		},
	}
	signaling_client.RegisterPeer(&signaling_client.Config{
		PeerId:   c.PeerId,
		Endpoint: c.SignalingEndpoint,
		LogLevel: c.LogLevel,
	}, handers)
}
