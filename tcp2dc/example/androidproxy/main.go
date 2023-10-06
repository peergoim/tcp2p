package main

import (
	"github.com/peergoim/tcp2p/tcp2dc"
	"github.com/peergoim/tcp2p/tcp2dc/internal/config"
)

func main() {
	tcp2dc.SetTcpServer(&config.Config{
		StunUrls:          nil,
		TargetPeerId:      "telegram",
		SignalingEndpoint: "http://localhost:31134",
		LogLevel:          "debug",
		ListenOn:          "0.0.0.0:10445",
	})
	select {}
}
