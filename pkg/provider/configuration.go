package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/tls"
	"maps"
	"reflect"
	"slices"
	"strings"
	"text/template"
	"unicode"
)

// Merge merges multiple configurations.
func Merge(ctx context.Context, configurations map[string]*dynamic.Configuration) *dynamic.Configuration {
	configuration := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores: make(map[string]tls.Store),
		},
	}

	servicesToDelete := map[string]struct{}{}
	services := map[string][]string{}

	routersToDelete := map[string]struct{}{}
	routers := map[string][]string{}

	servicesTCPToDelete := map[string]struct{}{}
	servicesTCP := map[string][]string{}

	routersTCPToDelete := map[string]struct{}{}
	routersTCP := map[string][]string{}

	servicesUDPToDelete := map[string]struct{}{}
	servicesUDP := map[string][]string{}

	routersUDPToDelete := map[string]struct{}{}
	routersUDP := map[string][]string{}

	middlewaresToDelete := map[string]struct{}{}
	middlewares := map[string][]string{}

	middlewaresTCPToDelete := map[string]struct{}{}
	middlewaresTCP := map[string][]string{}

	transportsToDelete := map[string]struct{}{}
	transports := map[string][]string{}

	transportsTCPToDelete := map[string]struct{}{}
	transportsTCP := map[string][]string{}

	storesToDelete := map[string]struct{}{}
	stores := map[string][]string{}

	var sortedKeys []string
	for key := range configurations {
		sortedKeys = append(sortedKeys, key)
	}
	slices.Sort(sortedKeys)

	for _, root := range sortedKeys {
		conf := configurations[root]
		for serviceName, service := range conf.HTTP.Services {
			services[serviceName] = append(services[serviceName], root)
			if !AddService(configuration.HTTP, serviceName, service) {
				servicesToDelete[serviceName] = struct{}{}
			}
		}

		for routerName, router := range conf.HTTP.Routers {
			routers[routerName] = append(routers[routerName], root)
			if !AddRouter(configuration.HTTP, routerName, router) {
				routersToDelete[routerName] = struct{}{}
			}
		}

		for transportName, transport := range conf.HTTP.ServersTransports {
			transports[transportName] = append(transports[transportName], root)
			if !AddTransport(configuration.HTTP, transportName, transport) {
				transportsToDelete[transportName] = struct{}{}
			}
		}

		for serviceName, service := range conf.TCP.Services {
			servicesTCP[serviceName] = append(servicesTCP[serviceName], root)
			if !AddServiceTCP(configuration.TCP, serviceName, service) {
				servicesTCPToDelete[serviceName] = struct{}{}
			}
		}

		for routerName, router := range conf.TCP.Routers {
			routersTCP[routerName] = append(routersTCP[routerName], root)
			if !AddRouterTCP(configuration.TCP, routerName, router) {
				routersTCPToDelete[routerName] = struct{}{}
			}
		}

		for transportName, transport := range conf.TCP.ServersTransports {
			transportsTCP[transportName] = append(transportsTCP[transportName], root)
			if !AddTransportTCP(configuration.TCP, transportName, transport) {
				transportsTCPToDelete[transportName] = struct{}{}
			}
		}

		for serviceName, service := range conf.UDP.Services {
			servicesUDP[serviceName] = append(servicesUDP[serviceName], root)
			if !AddServiceUDP(configuration.UDP, serviceName, service) {
				servicesUDPToDelete[serviceName] = struct{}{}
			}
		}

		for routerName, router := range conf.UDP.Routers {
			routersUDP[routerName] = append(routersUDP[routerName], root)
			if !AddRouterUDP(configuration.UDP, routerName, router) {
				routersUDPToDelete[routerName] = struct{}{}
			}
		}

		for middlewareName, middleware := range conf.HTTP.Middlewares {
			middlewares[middlewareName] = append(middlewares[middlewareName], root)
			if !AddMiddleware(configuration.HTTP, middlewareName, middleware) {
				middlewaresToDelete[middlewareName] = struct{}{}
			}
		}

		for middlewareName, middleware := range conf.TCP.Middlewares {
			middlewaresTCP[middlewareName] = append(middlewaresTCP[middlewareName], root)
			if !AddMiddlewareTCP(configuration.TCP, middlewareName, middleware) {
				middlewaresTCPToDelete[middlewareName] = struct{}{}
			}
		}

		for storeName, store := range conf.TLS.Stores {
			stores[storeName] = append(stores[storeName], root)
			if !AddStore(configuration.TLS, storeName, store) {
				storesToDelete[storeName] = struct{}{}
			}
		}
	}

	for serviceName := range servicesToDelete {
		log.Log.Error(log.ServiceName, serviceName,
			"configuration", services[serviceName],
			"Service defined multiple times with different configurations")
		delete(configuration.HTTP.Services, serviceName)
	}

	for routerName := range routersToDelete {
		log.Log.Error(log.RouterName, routerName,
			"configuration", routers[routerName],
			"Router defined multiple times with different configurations")
		delete(configuration.HTTP.Routers, routerName)
	}

	for transportName := range transportsToDelete {
		log.Log.Error(log.ServersTransportName, transportName,
			"configuration", transports[transportName],
			"ServersTransport defined multiple times with different configurations")
		delete(configuration.HTTP.ServersTransports, transportName)
	}

	for serviceName := range servicesTCPToDelete {
		log.Log.Error(log.ServiceName, serviceName,
			"configuration", servicesTCP[serviceName],
			"Service TCP defined multiple times with different configurations")
		delete(configuration.TCP.Services, serviceName)
	}

	for routerName := range routersTCPToDelete {
		log.Log.Error(log.RouterName, routerName,
			"configuration", routersTCP[routerName],
			"Router TCP defined multiple times with different configurations")
		delete(configuration.TCP.Routers, routerName)
	}

	for transportName := range transportsTCPToDelete {
		log.Log.Error(log.ServersTransportName, transportName,
			"configuration", transportsTCP[transportName],
			"ServersTransport TCP defined multiple times with different configurations")
		delete(configuration.TCP.ServersTransports, transportName)
	}

	for serviceName := range servicesUDPToDelete {
		log.Log.Error(log.ServiceName, serviceName,
			"configuration", servicesUDP[serviceName],
			"UDP service defined multiple times with different configurations")
		delete(configuration.UDP.Services, serviceName)
	}

	for routerName := range routersUDPToDelete {
		log.Log.Error(log.RouterName, routerName,
			"configuration", routersUDP[routerName],
			"UDP router defined multiple times with different configurations")
		delete(configuration.UDP.Routers, routerName)
	}

	for middlewareName := range middlewaresToDelete {
		log.Log.Error(log.MiddlewareName, middlewareName,
			"configuration", middlewares[middlewareName],
			"Middleware defined multiple times with different configurations")
		delete(configuration.HTTP.Middlewares, middlewareName)
	}

	for middlewareName := range middlewaresTCPToDelete {
		log.Log.Error(log.MiddlewareName, middlewareName,
			"configuration", middlewaresTCP[middlewareName],
			"TCP Middleware defined multiple times with different configurations")
		delete(configuration.TCP.Middlewares, middlewareName)
	}

	for storeName := range storesToDelete {
		log.Log.Errorf(fmt.Sprint("storeName", storeName,
			"TLS store defined multiple times with different configurations in %v"), stores[storeName])
		delete(configuration.TLS.Stores, storeName)
	}

	return configuration
}

// AddServiceTCP adds a service to a configuration.
func AddServiceTCP(configuration *dynamic.TCPConfiguration, serviceName string, service *dynamic.TCPService) bool {
	if _, ok := configuration.Services[serviceName]; !ok {
		configuration.Services[serviceName] = service
		return true
	}

	if !configuration.Services[serviceName].LoadBalancer.Mergeable(service.LoadBalancer) {
		return false
	}

	uniq := map[string]struct{}{}
	for _, server := range configuration.Services[serviceName].LoadBalancer.Servers {
		uniq[server.Address] = struct{}{}
	}

	for _, server := range service.LoadBalancer.Servers {
		if _, ok := uniq[server.Address]; !ok {
			configuration.Services[serviceName].LoadBalancer.Servers = append(configuration.Services[serviceName].LoadBalancer.Servers, server)
		}
	}

	return true
}

// AddRouterTCP adds a router to a configuration.
func AddRouterTCP(configuration *dynamic.TCPConfiguration, routerName string, router *dynamic.TCPRouter) bool {
	if _, ok := configuration.Routers[routerName]; !ok {
		configuration.Routers[routerName] = router
		return true
	}

	return reflect.DeepEqual(configuration.Routers[routerName], router)
}

// AddMiddlewareTCP adds a middleware to a configuration.
func AddMiddlewareTCP(configuration *dynamic.TCPConfiguration, middlewareName string, middleware *dynamic.TCPMiddleware) bool {
	if _, ok := configuration.Middlewares[middlewareName]; !ok {
		configuration.Middlewares[middlewareName] = middleware
		return true
	}

	return reflect.DeepEqual(configuration.Middlewares[middlewareName], middleware)
}

// AddTransportTCP adds a servers transport to a configuration.
func AddTransportTCP(configuration *dynamic.TCPConfiguration, transportName string, transport *dynamic.TCPServersTransport) bool {
	if _, ok := configuration.ServersTransports[transportName]; !ok {
		configuration.ServersTransports[transportName] = transport
		return true
	}

	return reflect.DeepEqual(configuration.ServersTransports[transportName], transport)
}

// AddServiceUDP adds a service to a configuration.
func AddServiceUDP(configuration *dynamic.UDPConfiguration, serviceName string, service *dynamic.UDPService) bool {
	if _, ok := configuration.Services[serviceName]; !ok {
		configuration.Services[serviceName] = service
		return true
	}

	if !configuration.Services[serviceName].LoadBalancer.Mergeable(service.LoadBalancer) {
		return false
	}

	uniq := map[string]struct{}{}
	for _, server := range configuration.Services[serviceName].LoadBalancer.Servers {
		uniq[server.Address] = struct{}{}
	}

	for _, server := range service.LoadBalancer.Servers {
		if _, ok := uniq[server.Address]; !ok {
			configuration.Services[serviceName].LoadBalancer.Servers = append(configuration.Services[serviceName].LoadBalancer.Servers, server)
		}
	}

	return true
}

// AddRouterUDP adds a router to a configuration.
func AddRouterUDP(configuration *dynamic.UDPConfiguration, routerName string, router *dynamic.UDPRouter) bool {
	if _, ok := configuration.Routers[routerName]; !ok {
		configuration.Routers[routerName] = router
		return true
	}

	return reflect.DeepEqual(configuration.Routers[routerName], router)
}

// AddService adds a service to a configuration.
func AddService(configuration *dynamic.HTTPConfiguration, serviceName string, service *dynamic.Service) bool {
	if _, ok := configuration.Services[serviceName]; !ok {
		configuration.Services[serviceName] = service
		return true
	}

	if !configuration.Services[serviceName].LoadBalancer.Mergeable(service.LoadBalancer) {
		return false
	}

	uniq := map[string]struct{}{}
	for _, server := range configuration.Services[serviceName].LoadBalancer.Servers {
		uniq[server.URL] = struct{}{}
	}

	for _, server := range service.LoadBalancer.Servers {
		if _, ok := uniq[server.URL]; !ok {
			configuration.Services[serviceName].LoadBalancer.Servers = append(configuration.Services[serviceName].LoadBalancer.Servers, server)
		}
	}

	return true
}

// AddRouter adds a router to a configuration.
func AddRouter(configuration *dynamic.HTTPConfiguration, routerName string, router *dynamic.Router) bool {
	if _, ok := configuration.Routers[routerName]; !ok {
		configuration.Routers[routerName] = router
		return true
	}

	return reflect.DeepEqual(configuration.Routers[routerName], router)
}

// AddTransport adds a servers transport to a configuration.
func AddTransport(configuration *dynamic.HTTPConfiguration, transportName string, transport *dynamic.ServersTransport) bool {
	if _, ok := configuration.ServersTransports[transportName]; !ok {
		configuration.ServersTransports[transportName] = transport
		return true
	}

	return reflect.DeepEqual(configuration.ServersTransports[transportName], transport)
}

// AddMiddleware adds a middleware to a configuration.
func AddMiddleware(configuration *dynamic.HTTPConfiguration, middlewareName string, middleware *dynamic.Middleware) bool {
	if _, ok := configuration.Middlewares[middlewareName]; !ok {
		configuration.Middlewares[middlewareName] = middleware
		return true
	}

	return reflect.DeepEqual(configuration.Middlewares[middlewareName], middleware)
}

// AddStore adds a middleware to a configurations.
func AddStore(configuration *dynamic.TLSConfiguration, storeName string, store tls.Store) bool {
	if _, ok := configuration.Stores[storeName]; !ok {
		configuration.Stores[storeName] = store
		return true
	}

	return reflect.DeepEqual(configuration.Stores[storeName], store)
}

// MakeDefaultRuleTemplate creates the default rule template.
func MakeDefaultRuleTemplate(defaultRule string, funcMap template.FuncMap) (*template.Template, error) {
	defaultFuncMap := sprig.TxtFuncMap()
	defaultFuncMap["normalize"] = Normalize

	for k, fn := range funcMap {
		defaultFuncMap[k] = fn
	}

	return template.New("defaultRule").Funcs(defaultFuncMap).Parse(defaultRule)
}

// BuildTCPRouterConfiguration builds a router configuration.
func BuildTCPRouterConfiguration(ctx context.Context, configuration *dynamic.TCPConfiguration) {
	for routerName, router := range configuration.Routers {
		var logRouter = log.Log.WithFields(map[string]interface{}{log.RouterName: routerName})
		if len(router.Rule) == 0 {
			delete(configuration.Routers, routerName)
			logRouter.Error("Empty rule")
			continue
		}

		if router.Service == "" {
			if len(configuration.Services) > 1 {
				delete(configuration.Routers, routerName)
				logRouter.Errorf("Router %s cannot be linked automatically with multiple Services: %q", routerName, slices.Collect(maps.Keys(configuration.Services)))
				continue
			}

			for serviceName := range configuration.Services {
				router.Service = serviceName
			}
		}
	}
}

// BuildUDPRouterConfiguration builds a router configuration.
func BuildUDPRouterConfiguration(ctx context.Context, configuration *dynamic.UDPConfiguration) {
	for routerName, router := range configuration.Routers {
		var logRouter = log.Log.WithFields(map[string]interface{}{log.RouterName: routerName})
		if router.Service != "" {
			continue
		}

		if len(configuration.Services) > 1 {
			delete(configuration.Routers, routerName)
			logRouter.Errorf("Router %s cannot be linked automatically with multiple Services: %q", routerName, slices.Collect(maps.Keys(configuration.Services)))
			continue
		}

		for serviceName := range configuration.Services {
			router.Service = serviceName
			break
		}
	}
}

// BuildRouterConfiguration builds a router configuration.
func BuildRouterConfiguration(ctx context.Context, configuration *dynamic.HTTPConfiguration, defaultRouterName string, defaultRuleTpl *template.Template, model interface{}) {
	if len(configuration.Routers) == 0 {
		if len(configuration.Services) > 1 {
			log.Log.Info("Could not create a router for the container: too many services")
		} else {
			configuration.Routers = make(map[string]*dynamic.Router)
			configuration.Routers[defaultRouterName] = &dynamic.Router{}
		}
	}

	for routerName, router := range configuration.Routers {
		var logRouter = log.Log.WithFields(map[string]interface{}{log.RouterName: routerName})

		if len(router.Rule) == 0 {
			writer := &bytes.Buffer{}
			if err := defaultRuleTpl.Execute(writer, model); err != nil {
				logRouter.WithError(err).Error("Error while parsing default rule")
				delete(configuration.Routers, routerName)
				continue
			}

			router.Rule = writer.String()
			if len(router.Rule) == 0 {
				logRouter.Error("Undefined rule")
				delete(configuration.Routers, routerName)
				continue
			}

			// Flag default rule routers to add the denyRouterRecursion middleware.
			router.DefaultRule = true
		}

		if router.Service == "" {
			if len(configuration.Services) > 1 {
				delete(configuration.Routers, routerName)
				log.Log.Errorf("Router %s cannot be linked automatically with multiple Services: %q", routerName, slices.Collect(maps.Keys(configuration.Services)))
				continue
			}

			for serviceName := range configuration.Services {
				router.Service = serviceName
			}
		}
	}
}

// Normalize replaces all special chars with `-`.
func Normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}
