package config

import (
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/poxyResource/provider"
)

func mergeConfiguration(configurations dynamic.Configurations) dynamic.Configuration {
	// TODO: see if we can use DeepCopies inside, so that the given argument is left
	// untouched, and the modified copy is returned.
	conf := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services: make(map[string]*dynamic.Service),
		},
	}

	for pvd, configuration := range configurations {
		if configuration.HTTP != nil {
			for serviceName, service := range configuration.HTTP.Services {
				conf.HTTP.Services[provider.MakeQualifiedName(pvd, serviceName)] = service
			}
		}
	}
	return conf
}
