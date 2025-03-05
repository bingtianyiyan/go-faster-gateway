package api

import (
	"go-faster-gateway/pkg/config/static"
	ptypes "go-faster-gateway/pkg/types"
	"time"
)

// GatewayCmdConfiguration wraps the static configuration and extra parameters.
type GatewayCmdConfiguration struct {
	static.Configuration `export:"true"`
	// ConfigFile is the path to the configuration file.
	ConfigFile string `description:"Configuration file to use. If specified all other flags are ignored." export:"true"`
}

// NewGatewayConfiguration creates a GatewayCmdConfiguration with default values.
func NewGatewayConfiguration() *GatewayCmdConfiguration {
	return &GatewayCmdConfiguration{
		Configuration: static.Configuration{
			EntryPoints: make(static.EntryPoints),
			Providers: &static.Providers{
				ProvidersThrottleDuration: ptypes.Duration(2 * time.Second),
			},
			ServersTransport: &static.ServersTransport{
				MaxIdleConnsPerHost: 200,
			},
			TCPServersTransport: &static.TCPServersTransport{
				DialTimeout:   ptypes.Duration(30 * time.Second),
				DialKeepAlive: ptypes.Duration(15 * time.Second),
			},
		},
		ConfigFile: "",
	}
}
