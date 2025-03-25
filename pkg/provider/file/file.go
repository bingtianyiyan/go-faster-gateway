package file

import (
	"context"
	"errors"
	"fmt"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/utils"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider"
	"go-faster-gateway/pkg/safe"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

const providerName = "file"

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	Watch    bool   `description:"Watch provider." json:"watch,omitempty" toml:"watch,omitempty" yaml:"watch,omitempty" export:"true"`
	Filename string `description:"Load dynamic configuration from a file." json:"filename,omitempty" toml:"filename,omitempty" yaml:"filename,omitempty" export:"true"`
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

// Provide allows the file provider to provide configurations to
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	slog := log.Log.WithFields(map[string]interface{}{log.ProviderName: providerName})

	if p.Watch {
		var watchItems []string

		switch {
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
					slog.Error(zap.Error(err), "Error while building configuration")
				}
			}
		}
	})

	if err := p.applyConfiguration(configurationChan); err != nil {
		if p.Watch {
			slog.Error(zap.Error(err), "Error while building configuration (for the first time)")
			return nil
		}

		return err
	}

	return nil
}

// GetConfig for provider get config
func (p *Provider) GetConfig() (dynamic.Message, error) {
	configuration, err := p.buildConfiguration()
	return dynamic.Message{
		ProviderName:  "file",
		Configuration: configuration,
	}, err
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
		slog := log.Log.WithFields(map[string]interface{}{log.ProviderName: providerName})
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-watcher.Events:
				if len(p.Filename) > 0 {
					_, evtFileName := filepath.Split(evt.Name)
					_, confFileName := filepath.Split(p.Filename)
					if evtFileName == confFileName {
						err := callback(configurationChan)
						if err != nil {
							slog.WithError(err).Error("Error occurred during watcher callback")
						}
					}
				} else {
					err := callback(configurationChan)
					if err != nil {
						slog.WithError(err).Error("Error occurred during watcher callback")
					}
				}
			case err := <-watcher.Errors:
				slog.WithError(err).Error("Watcher event error")
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

// buildConfiguration loads configuration either from file
// specified by 'Filename' and returns a 'Configuration' object.
func (p *Provider) buildConfiguration() (*dynamic.Configuration, error) {
	if len(p.Filename) > 0 {
		return p.loadFileConfig(p.Filename)
	}
	return nil, errors.New("error using file configuration provider, neither filename is not defined")
}

func sendConfigToChannel(configurationChan chan<- dynamic.Message, configuration *dynamic.Configuration) {
	configurationChan <- dynamic.Message{
		ProviderName:  "file",
		Configuration: configuration,
	}
}

func (p *Provider) loadFileConfig(filename string) (*dynamic.Configuration, error) {
	c := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Services: make(map[string]*dynamic.Service),
		},
	}
	err := utils.GetFile(filename, c)
	return c, err
}
