package provider

import (
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/safe"
)

// Provider defines methods of a provider.
type Provider interface {
	// Provide allows the provider to provide configurations to gateway
	// using the given configuration channel.
	Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error
	Init() error
	GetConfig() (dynamic.Message, error)
}
