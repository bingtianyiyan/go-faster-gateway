package config

import (
	"context"
	"encoding/json"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/utils"
	"go-faster-gateway/pkg/log"
	logger2 "go-faster-gateway/pkg/log/logger"
	"go-faster-gateway/pkg/provider"
	"go-faster-gateway/pkg/safe"
	"reflect"
)

// ConfigurationWatcher watches configuration changes.
type ConfigurationWatcher struct {
	providerAggregator provider.Provider

	//defaultEntryPoints []string

	allProvidersConfigs chan dynamic.Message

	newConfigs chan dynamic.Configurations

	requiredProvider       string
	configurationListeners []func(dynamic.Configuration)

	routinesPool *safe.Pool
}

// NewConfigurationWatcher creates a new ConfigurationWatcher.
func NewConfigurationWatcher(
	routinesPool *safe.Pool,
	pvd provider.Provider,
	requiredProvider string,
) *ConfigurationWatcher {
	return &ConfigurationWatcher{
		providerAggregator:  pvd,
		allProvidersConfigs: make(chan dynamic.Message, 100),
		newConfigs:          make(chan dynamic.Configurations),
		routinesPool:        routinesPool,
		requiredProvider:    requiredProvider,
	}
}

// Start the configuration watcher.
func (c *ConfigurationWatcher) Start() {
	c.routinesPool.GoCtx(c.receiveConfigurations)
	c.routinesPool.GoCtx(c.applyConfigurations)
	//读取配置文件
	c.startProviderAggregator()
}

// Stop the configuration watcher.
func (c *ConfigurationWatcher) Stop() {
	close(c.allProvidersConfigs)
	close(c.newConfigs)
}

// AddListener adds a new listener function used when new configuration is provided.
func (c *ConfigurationWatcher) AddListener(listener func(dynamic.Configuration)) {
	if c.configurationListeners == nil {
		c.configurationListeners = make([]func(dynamic.Configuration), 0)
	}
	c.configurationListeners = append(c.configurationListeners, listener)
}

func (c *ConfigurationWatcher) startProviderAggregator() {
	log.Log.Infof("Starting provider aggregator %T", c.providerAggregator)

	safe.Go(func() {
		err := c.providerAggregator.Provide(c.allProvidersConfigs, c.routinesPool)
		if err != nil {
			log.Log.WithError(err).Errorf("Error starting provider aggregator %T", c.providerAggregator)
		}
	})
}

// receiveConfigurations receives configuration changes from the providers.
// The configuration message then gets passed along a series of check, notably
// to verify that, for a given provider, the configuration that was just received
// is at least different from the previously received one.
// The full set of configurations is then sent to the throttling goroutine,
// (throttleAndApplyConfigurations) via a RingChannel, which ensures that we can
// constantly send in a non-blocking way to the throttling goroutine the last
// global state we are aware of.
func (c *ConfigurationWatcher) receiveConfigurations(ctx context.Context) {
	ctx = logger2.NewContext(ctx, log.Log)
	newConfigurations := make(dynamic.Configurations)
	var output chan dynamic.Configurations
	for {
		select {
		case <-ctx.Done():
			return
		// DeepCopy is necessary because newConfigurations gets modified later by the consumer of c.newConfigs
		case output <- *utils.DeepCopyDefault(nil, &newConfigurations):
			output = nil

		default:
			select {
			case <-ctx.Done():
				return
			case configMsg, ok := <-c.allProvidersConfigs:
				if !ok {
					return
				}
				slog, _ := logger2.FromContext(ctx)
				slog.WithFields(map[string]interface{}{
					log.ProviderName: configMsg.ProviderName})

				if configMsg.Configuration == nil {
					slog.Debug("Skipping nil configuration")
					continue
				}

				if isEmptyConfiguration(configMsg.Configuration) {
					slog.Debug("Skipping empty configuration")
					continue
				}

				logConfiguration(slog, configMsg)
				//	dynamic.Configuration
				dyConfig, ok := newConfigurations[configMsg.ProviderName]
				if ok {
					if reflect.DeepEqual(dyConfig, configMsg.Configuration) {
						// no change, do nothing
						slog.Debug("Skipping unchanged configuration")
						continue
					}
				}

				newConfigurations[configMsg.ProviderName], _ = utils.DeepCopy(nil, configMsg.Configuration)
				output = c.newConfigs

			// DeepCopy is necessary because newConfigurations gets modified later by the consumer of c.newConfigs
			case output <- *utils.DeepCopyDefault(nil, &newConfigurations):
				output = nil
			}
		}
	}
}

// applyConfigurations blocks on a RingChannel that receives the new
// set of configurations that is compiled and sent by receiveConfigurations as soon
// as a provider change occurs. If the new set is different from the previous set
// that had been applied, the new set is applied, and we sleep for a while before
// listening on the channel again.
func (c *ConfigurationWatcher) applyConfigurations(ctx context.Context) {
	var lastConfigurations dynamic.Configurations
	for {
		select {
		case <-ctx.Done():
			return
		case newConfigs, ok := <-c.newConfigs:
			if !ok {
				return
			}

			// We wait for first configuration of the required provider before applying configurations.
			if _, ok := newConfigs[c.requiredProvider]; c.requiredProvider != "" && !ok {
				continue
			}

			if reflect.DeepEqual(newConfigs, lastConfigurations) {
				continue
			}

			conf, err := utils.DeepCopyMap(nil, newConfigs)
			if err == nil {
				for _, listener := range c.configurationListeners {
					listener(*conf)
				}
			}

			lastConfigurations = newConfigs
			log.Log.Debugf("lastConfiguration is %s", utils.JsonMarshalToStrNoErr(lastConfigurations))
		}
	}
}

func logConfiguration(slog *logger2.Helper, configMsg dynamic.Message) {
	copyConf, err := utils.DeepCopy(nil, configMsg.Configuration)
	jsonConf, err := json.Marshal(copyConf)
	if err != nil {
		slog.WithError(err).Error("Could not marshal dynamic configuration")
		slog.Debugf("Configuration received: [struct] %#v", copyConf)
	} else {
		slog.Debugf("Configuration received %s", string(jsonConf))
	}
}

func isEmptyConfiguration(conf *dynamic.Configuration) bool {
	if conf.HTTP == nil {
		conf.HTTP = &dynamic.HTTPConfiguration{
			Services: make(map[string]*dynamic.Service),
		}
	}

	httpEmpty := conf.HTTP.Services == nil
	return httpEmpty
}
