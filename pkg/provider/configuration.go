package provider

import (
	"context"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/log/logger"
	"slices"
	"strings"
	"unicode"
)

// Merge merges multiple configurations.
func Merge(ctx context.Context, configurations map[string]*dynamic.Configuration) *dynamic.Configuration {
	slog, _ := logger.FromContext(ctx)
	configuration := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services: make(map[string]*dynamic.Service),
		},
	}
	var sortedKeys []string
	for key := range configurations {
		sortedKeys = append(sortedKeys, key)
	}
	slices.Sort(sortedKeys)

	servicesToDelete := map[string]struct{}{}
	services := map[string][]string{}
	for _, root := range sortedKeys {
		conf := configurations[root]
		for serviceName, service := range conf.HTTP.Services {
			services[serviceName] = append(services[serviceName], root)
			if !AddService(configuration.HTTP, serviceName, service) {
				servicesToDelete[serviceName] = struct{}{}
			}
		}
	}
	for serviceName := range servicesToDelete {
		slog.WithFields(map[string]interface{}{
			log.ServiceName: serviceName,
			"configuration": services[serviceName],
		}).Error("Service defined multiple times with different configurations")
		delete(configuration.HTTP.Services, serviceName)
	}
	return configuration
}

// AddService adds a service to a configuration.
func AddService(configuration *dynamic.HTTPConfiguration, serviceName string, service *dynamic.Service) bool {
	if _, ok := configuration.Services[serviceName]; !ok {
		configuration.Services[serviceName] = service
		return true
	}
	return true
}

// Normalize replaces all special chars with `-`.
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}
