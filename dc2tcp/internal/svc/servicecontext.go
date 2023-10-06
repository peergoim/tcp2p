package svc

import "github.com/peergoim/tcp2p/dc2tcp/internal/config"

type ServiceContext struct {
	Config *config.Config
}

func NewServiceContext(config *config.Config) *ServiceContext {
	return &ServiceContext{Config: config}
}
