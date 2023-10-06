package tcp2dc

import (
	"github.com/peergoim/tcp2p/tcp2dc/internal/config"
	"github.com/peergoim/tcp2p/tcp2dc/internal/server"
	"strings"
)

func SetTcpServer(c *config.Config) error {
	if len(c.StunUrls) == 0 {
		c.StunUrls = config.DefaultStunUrls
	}
	if !strings.HasPrefix(c.SignalingEndpoint, "http") {
		panic("signaling endpoint must start with http or https")
	}
	if c.ListenOn == "" {
		panic("listen on is empty")
	} else if len(strings.Split(c.ListenOn, ":")) != 2 {
		panic("listen on is invalid")
	}
	if c.TargetPeerId == "" {
		panic("target peer id is empty")
	}
	return server.SetTcpServer(c)
}
