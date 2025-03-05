package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/file"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/log/logger"
	"go-faster-gateway/pkg/provider"
	"go-faster-gateway/pkg/safe"
	"go-faster-gateway/pkg/tls"
	"go-faster-gateway/pkg/types"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/fsnotify/fsnotify"
)

const providerName = "file"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Directory                 string `description:"Load dynamic configuration from one or more .yml or .toml files in a directory." json:"directory,omitempty" toml:"directory,omitempty" yaml:"directory,omitempty" export:"true"`
	Watch                     bool   `description:"Watch provider." json:"watch,omitempty" toml:"watch,omitempty" yaml:"watch,omitempty" export:"true"`
	Filename                  string `description:"Load dynamic configuration from a file." json:"filename,omitempty" toml:"filename,omitempty" yaml:"filename,omitempty" export:"true"`
	DebugLogGeneratedTemplate bool   `description:"Enable debug logging of generated configuration template." json:"debugLogGeneratedTemplate,omitempty" toml:"debugLogGeneratedTemplate,omitempty" yaml:"debugLogGeneratedTemplate,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Watch = true
	p.Filename = ""
}

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Provide allows the file provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	logger := log.Log.WithFields(map[string]interface{}{log.ProviderName: providerName})

	if p.Watch {
		var watchItems []string

		switch {
		case len(p.Directory) > 0:
			watchItems = append(watchItems, p.Directory)

			fileList, err := os.ReadDir(p.Directory)
			if err != nil {
				return fmt.Errorf("unable to read directory %s: %w", p.Directory, err)
			}

			for _, entry := range fileList {
				if entry.IsDir() {
					// ignore sub-dir
					continue
				}
				watchItems = append(watchItems, path.Join(p.Directory, entry.Name()))
			}
		case len(p.Filename) > 0:
			watchItems = append(watchItems, filepath.Dir(p.Filename), p.Filename)
		default:
			return errors.New("error using file configuration provider, neither filename nor directory is defined")
		}

		if err := p.addWatcher(pool, watchItems, configurationChan, p.applyConfiguration); err != nil {
			return err
		}
	}

	pool.GoCtx(func(ctx context.Context) {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGHUP)

		for {
			select {
			case <-ctx.Done():
				return
			// signals only receives SIGHUP events.
			case <-signals:
				if err := p.applyConfiguration(configurationChan); err != nil {
					logger.Error(zap.Error(err), "Error while building configuration")
				}
			}
		}
	})

	if err := p.applyConfiguration(configurationChan); err != nil {
		if p.Watch {
			logger.Error(zap.Error(err), "Error while building configuration (for the first time)")
			return nil
		}

		return err
	}

	return nil
}

func (p *Provider) addWatcher(pool *safe.Pool, items []string, configurationChan chan<- dynamic.Message, callback func(chan<- dynamic.Message) error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating file watcher: %w", err)
	}

	for _, item := range items {
		log.Log.Debugf("add watcher on: %s", item)
		err = watcher.Add(item)
		if err != nil {
			return fmt.Errorf("error adding file watcher: %w", err)
		}
	}

	// Process events
	pool.GoCtx(func(ctx context.Context) {
		logger := log.Log.WithFields(map[string]interface{}{log.ProviderName: providerName})
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-watcher.Events:
				if p.Directory == "" {
					_, evtFileName := filepath.Split(evt.Name)
					_, confFileName := filepath.Split(p.Filename)
					if evtFileName == confFileName {
						err := callback(configurationChan)
						if err != nil {
							logger.WithError(err).Error("Error occurred during watcher callback")
						}
					}
				} else {
					err := callback(configurationChan)
					if err != nil {
						logger.WithError(err).Error("Error occurred during watcher callback")
					}
				}
			case err := <-watcher.Errors:
				logger.WithError(err).Error("Watcher event error")
			}
		}
	})
	return nil
}

// applyConfiguration builds the configuration and sends it to the given configurationChan.
func (p *Provider) applyConfiguration(configurationChan chan<- dynamic.Message) error {
	configuration, err := p.buildConfiguration()
	if err != nil {
		return err
	}

	sendConfigToChannel(configurationChan, configuration)
	return nil
}

// buildConfiguration loads configuration either from file or a directory
// specified by 'Filename'/'Directory' and returns a 'Configuration' object.
func (p *Provider) buildConfiguration() (*dynamic.Configuration, error) {
	var log = log.Log.WithFields(map[string]interface{}{log.ProviderName: providerName})
	ctx := logger.NewContext(context.Background(), log)

	if len(p.Directory) > 0 {
		return p.loadFileConfigFromDirectory(ctx, p.Directory, nil)
	}

	if len(p.Filename) > 0 {
		return p.loadFileConfig(ctx, p.Filename, true)
	}

	return nil, errors.New("error using file configuration provider, neither filename nor directory is defined")
}

func sendConfigToChannel(configurationChan chan<- dynamic.Message, configuration *dynamic.Configuration) {
	configurationChan <- dynamic.Message{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

func (p *Provider) loadFileConfig(ctx context.Context, filename string, parseTemplate bool) (*dynamic.Configuration, error) {
	var err error
	var configuration *dynamic.Configuration
	if parseTemplate {
		configuration, err = p.CreateConfiguration(ctx, filename, template.FuncMap{}, false)
	} else {
		configuration, err = p.DecodeConfiguration(filename)
	}
	if err != nil {
		return nil, err
	}

	if configuration.TLS != nil {
		configuration.TLS.Certificates = flattenCertificates(ctx, configuration.TLS)

		// TLS Options
		if configuration.TLS.Options != nil {
			for name, options := range configuration.TLS.Options {
				var caCerts []types.FileOrContent

				for _, caFile := range options.ClientAuth.CAFiles {
					content, err := caFile.Read()
					if err != nil {
						log, _ := logger.FromContext(ctx)
						log.WithError(err)
						continue
					}

					caCerts = append(caCerts, types.FileOrContent(content))
				}
				options.ClientAuth.CAFiles = caCerts

				configuration.TLS.Options[name] = options
			}
		}

		// TLS stores
		if len(configuration.TLS.Stores) > 0 {
			for name, store := range configuration.TLS.Stores {
				if store.DefaultCertificate == nil {
					continue
				}

				content, err := store.DefaultCertificate.CertFile.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}
				store.DefaultCertificate.CertFile = types.FileOrContent(content)

				content, err = store.DefaultCertificate.KeyFile.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}
				store.DefaultCertificate.KeyFile = types.FileOrContent(content)

				configuration.TLS.Stores[name] = store
			}
		}
	}

	// HTTP ServersTransport
	if configuration.HTTP != nil && len(configuration.HTTP.ServersTransports) > 0 {
		for name, st := range configuration.HTTP.ServersTransports {
			var certificates []tls.Certificate
			for _, cert := range st.Certificates {
				content, err := cert.CertFile.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}
				cert.CertFile = types.FileOrContent(content)

				content, err = cert.KeyFile.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}
				cert.KeyFile = types.FileOrContent(content)

				certificates = append(certificates, cert)
			}

			configuration.HTTP.ServersTransports[name].Certificates = certificates

			var rootCAs []types.FileOrContent
			for _, rootCA := range st.RootCAs {
				content, err := rootCA.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(content))
			}

			st.RootCAs = rootCAs
		}
	}

	// TCP ServersTransport
	if configuration.TCP != nil && len(configuration.TCP.ServersTransports) > 0 {
		for name, st := range configuration.TCP.ServersTransports {
			var certificates []tls.Certificate
			if st.TLS == nil {
				continue
			}
			for _, cert := range st.TLS.Certificates {
				content, err := cert.CertFile.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}
				cert.CertFile = types.FileOrContent(content)

				content, err = cert.KeyFile.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}
				cert.KeyFile = types.FileOrContent(content)

				certificates = append(certificates, cert)
			}

			configuration.TCP.ServersTransports[name].TLS.Certificates = certificates

			var rootCAs []types.FileOrContent
			for _, rootCA := range st.TLS.RootCAs {
				content, err := rootCA.Read()
				if err != nil {
					log, _ := logger.FromContext(ctx)
					log.WithError(err)
					continue
				}

				rootCAs = append(rootCAs, types.FileOrContent(content))
			}

			st.TLS.RootCAs = rootCAs
		}
	}

	return configuration, nil
}

func flattenCertificates(ctx context.Context, tlsConfig *dynamic.TLSConfiguration) []*tls.CertAndStores {
	var certs []*tls.CertAndStores
	for _, cert := range tlsConfig.Certificates {
		content, err := cert.Certificate.CertFile.Read()
		if err != nil {
			log, _ := logger.FromContext(ctx)
			log.WithError(err)
			continue
		}
		cert.Certificate.CertFile = types.FileOrContent(string(content))

		content, err = cert.Certificate.KeyFile.Read()
		if err != nil {
			log, _ := logger.FromContext(ctx)
			log.WithError(err)
			continue
		}
		cert.Certificate.KeyFile = types.FileOrContent(string(content))

		certs = append(certs, cert)
	}

	return certs
}

func (p *Provider) loadFileConfigFromDirectory(ctx context.Context, directory string, configuration *dynamic.Configuration) (*dynamic.Configuration, error) {
	fileList, err := os.ReadDir(directory)
	if err != nil {
		return configuration, fmt.Errorf("unable to read directory %s: %w", directory, err)
	}

	if configuration == nil {
		configuration = &dynamic.Configuration{
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
			TLS: &dynamic.TLSConfiguration{
				Stores:  make(map[string]tls.Store),
				Options: make(map[string]tls.Options),
			},
			UDP: &dynamic.UDPConfiguration{
				Routers:  make(map[string]*dynamic.UDPRouter),
				Services: make(map[string]*dynamic.UDPService),
			},
		}
	}

	configTLSMaps := make(map[*tls.CertAndStores]struct{})
	for _, item := range fileList {
		logTemp, _ := logger.FromContext(ctx)
		logTemp.WithFields(map[string]interface{}{"filename": item.Name()})
		if item.IsDir() {
			configuration, err = p.loadFileConfigFromDirectory(logger.NewContext(ctx, logTemp), filepath.Join(directory, item.Name()), configuration)
			if err != nil {
				return configuration, fmt.Errorf("unable to load content configuration from subdirectory %s: %w", item, err)
			}
			continue
		}

		switch strings.ToLower(filepath.Ext(item.Name())) {
		case ".toml", ".yaml", ".yml":
			// noop
		default:
			continue
		}

		var c *dynamic.Configuration
		c, err = p.loadFileConfig(logger.NewContext(ctx, logTemp), filepath.Join(directory, item.Name()), true)
		if err != nil {
			return configuration, fmt.Errorf("%s: %w", filepath.Join(directory, item.Name()), err)
		}

		for name, conf := range c.HTTP.Routers {
			if _, exists := configuration.HTTP.Routers[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.RouterName: name}).Warn("HTTP router already configured, skipping")
			} else {
				configuration.HTTP.Routers[name] = conf
			}
		}

		for name, conf := range c.HTTP.Middlewares {
			if _, exists := configuration.HTTP.Middlewares[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.MiddlewareName: name}).Warn("HTTP middleware already configured, skipping")
			} else {
				configuration.HTTP.Middlewares[name] = conf
			}
		}

		for name, conf := range c.HTTP.Services {
			if _, exists := configuration.HTTP.Services[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.ServiceName: name}).Warn("HTTP service already configured, skipping")
			} else {
				configuration.HTTP.Services[name] = conf
			}
		}

		for name, conf := range c.HTTP.ServersTransports {
			if _, exists := configuration.HTTP.ServersTransports[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.ServersTransportName: name}).Warn("HTTP servers transport already configured, skipping")
			} else {
				configuration.HTTP.ServersTransports[name] = conf
			}
		}

		for name, conf := range c.TCP.Routers {
			if _, exists := configuration.TCP.Routers[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.RouterName: name}).Warn("TCP router already configured, skipping")
			} else {
				configuration.TCP.Routers[name] = conf
			}
		}

		for name, conf := range c.TCP.Middlewares {
			if _, exists := configuration.TCP.Middlewares[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.MiddlewareName: name}).Warn("TCP middleware already configured, skipping")
			} else {
				configuration.TCP.Middlewares[name] = conf
			}
		}

		for name, conf := range c.TCP.Services {
			if _, exists := configuration.TCP.Services[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.ServiceName: name}).Warn("TCP service already configured, skipping")
			} else {
				configuration.TCP.Services[name] = conf
			}
		}

		for name, conf := range c.TCP.ServersTransports {
			if _, exists := configuration.TCP.ServersTransports[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.ServersTransportName: name}).Warn("TCP servers transport already configured, skipping")
			} else {
				configuration.TCP.ServersTransports[name] = conf
			}
		}

		for name, conf := range c.UDP.Routers {
			if _, exists := configuration.UDP.Routers[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.RouterName: name}).Warn("UDP router already configured, skipping")
			} else {
				configuration.UDP.Routers[name] = conf
			}
		}

		for name, conf := range c.UDP.Services {
			if _, exists := configuration.UDP.Services[name]; exists {
				log.Log.WithFields(map[string]interface{}{log.ServiceName: name}).Warn("UDP service already configured, skipping")
			} else {
				configuration.UDP.Services[name] = conf
			}
		}

		for _, conf := range c.TLS.Certificates {
			if _, exists := configTLSMaps[conf]; exists {
				log.Log.Warnf("TLS configuration %v already configured, skipping", conf)
			} else {
				configTLSMaps[conf] = struct{}{}
			}
		}

		for name, conf := range c.TLS.Options {
			if _, exists := configuration.TLS.Options[name]; exists {
				log.Log.Warnf("TLS options %v already configured, skipping", name)
			} else {
				if configuration.TLS.Options == nil {
					configuration.TLS.Options = map[string]tls.Options{}
				}
				configuration.TLS.Options[name] = conf
			}
		}

		for name, conf := range c.TLS.Stores {
			if _, exists := configuration.TLS.Stores[name]; exists {
				log.Log.Warnf("TLS store %v already configured, skipping", name)
			} else {
				if configuration.TLS.Stores == nil {
					configuration.TLS.Stores = map[string]tls.Store{}
				}
				configuration.TLS.Stores[name] = conf
			}
		}
	}

	if len(configTLSMaps) > 0 && configuration.TLS == nil {
		configuration.TLS = &dynamic.TLSConfiguration{}
	}

	for conf := range configTLSMaps {
		configuration.TLS.Certificates = append(configuration.TLS.Certificates, conf)
	}

	return configuration, nil
}

// CreateConfiguration creates a provider configuration from content using templating.
func (p *Provider) CreateConfiguration(ctx context.Context, filename string, funcMap template.FuncMap, templateObjects interface{}) (*dynamic.Configuration, error) {
	tmplContent, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %w", filename, err)
	}

	defaultFuncMap := sprig.TxtFuncMap()
	defaultFuncMap["normalize"] = provider.Normalize
	defaultFuncMap["split"] = strings.Split
	for funcID, funcElement := range funcMap {
		defaultFuncMap[funcID] = funcElement
	}

	tmpl := template.New(p.Filename).Funcs(defaultFuncMap)

	_, err = tmpl.Parse(tmplContent)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, templateObjects)
	if err != nil {
		return nil, err
	}

	renderedTemplate := buffer.String()
	if p.DebugLogGeneratedTemplate {
		log, _ := logger.FromContext(ctx)
		log.Debugf("Template content: %s", tmplContent)
		log.Debugf("Rendering results: %s", renderedTemplate)
	}

	return p.decodeConfiguration(filename, renderedTemplate)
}

// DecodeConfiguration Decodes a *types.Configuration from a content.
func (p *Provider) DecodeConfiguration(filename string) (*dynamic.Configuration, error) {
	content, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s - %w", filename, err)
	}

	return p.decodeConfiguration(filename, content)
}

func (p *Provider) decodeConfiguration(filePath, content string) (*dynamic.Configuration, error) {
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
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}

	err := file.DecodeContent(content, strings.ToLower(filepath.Ext(filePath)), configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func readFile(filename string) (string, error) {
	if len(filename) > 0 {
		buf, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	return "", fmt.Errorf("invalid filename: %s", filename)
}
