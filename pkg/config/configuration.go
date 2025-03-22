package config

import (
	"go-faster-gateway/pkg/config/static"
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
			Providers: &static.Providers{},
		},
		ConfigFile: "",
	}
}
