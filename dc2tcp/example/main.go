package main

import (
	"github.com/peergoim/tcp2p/dc2tcp"
	"github.com/peergoim/tcp2p/dc2tcp/internal/config"
)

func main() {
	dc2tcp.Run(&config.Config{
		StunUrls:          nil,
		TcpAddr:           "127.0.0.1:10444",
		PeerId:            "telegram",
		SignalingEndpoint: "ws://localhost:31134",
		LogLevel:          "debug",
	})
	select {}
}
