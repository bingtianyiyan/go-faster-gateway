package aggregator

import (
	"context"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/config/static"
	"go-faster-gateway/pkg/helper/utils"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider"
	"go-faster-gateway/pkg/provider/file"
	"go-faster-gateway/pkg/redactor"
	"go-faster-gateway/pkg/safe"
	"time"
)

// throttled defines what kind of config refresh throttling the aggregator should
// set up for a given provider.
// If a provider implements throttled, the configuration changes it sends will be
// taken into account no more often than the frequency inferred from ThrottleDuration().
// If ThrottleDuration returns zero, no throttling will take place.
// If throttled is not implemented, the throttling will be set up in accordance
// with the global providersThrottleDuration option.
// “throttled”是一个被定义或配置的属性，它指定了聚合器在处理配置刷新时应该采用的节流策略或类型。
// 节流（throttling）通常用于控制资源的使用率或限制操作的频率，以避免过载或滥用。
// 在配置聚合器的上下文中，这可能意味着对配置更新的频率、数量或速度进行限制，以确保聚合器能够稳定、高效地处理这些更新。
type throttled interface {
	ThrottleDuration() time.Duration
}

// maybeThrottledProvide returns the Provide method of the given provider,
// potentially augmented with some throttling depending on whether and how the
// provider implements the throttled interface.
func maybeThrottledProvide(prd provider.Provider, defaultDuration time.Duration) func(chan<- dynamic.Message, *safe.Pool) error {
	providerThrottleDuration := defaultDuration
	if throttled, ok := prd.(throttled); ok {
		// per-provider throttling
		providerThrottleDuration = throttled.ThrottleDuration()
	}

	if providerThrottleDuration == 0 {
		// throttling disabled
		return prd.Provide
	}

	return func(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
		rc := newRingChannel()
		pool.GoCtx(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-rc.out():
					configurationChan <- msg
					time.Sleep(providerThrottleDuration)
				}
			}
		})

		return prd.Provide(rc.in(), pool)
	}
}

// ProviderAggregator aggregates providers.
type ProviderAggregator struct {
	//程序默认以文件方式获取配置信息
	fileProvider              provider.Provider
	providers                 []provider.Provider
	providersThrottleDuration time.Duration
}

// NewProviderAggregator returns an aggregate of all the providers configured in the static configuration.
func NewProviderAggregator(conf static.Providers) *ProviderAggregator {
	p := &ProviderAggregator{
		providersThrottleDuration: time.Duration(conf.ProvidersThrottleDuration),
	}

	if conf.File != nil {
		p.quietAddProvider(conf.File)
	}

	//如果有其他类型提供配置文件，比如nacos,apollo等

	return p
}

func (p *ProviderAggregator) quietAddProvider(provider provider.Provider) {
	err := p.AddProvider(provider)
	if err != nil {
		log.Log.WithError(err).Errorf("Error while initializing provider %T", provider)
	}
}

// AddProvider adds a provider in the providers map.
func (p *ProviderAggregator) AddProvider(provider provider.Provider) error {
	err := provider.Init()
	if err != nil {
		return err
	}

	switch provider.(type) {
	case *file.Provider:
		p.fileProvider = provider
	default:
		p.providers = append(p.providers, provider)
	}

	return nil
}

// Init the provider.
func (p *ProviderAggregator) Init() error {
	return nil
}

// Provide calls the provide method of every providers.
func (p *ProviderAggregator) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	if p.fileProvider != nil {
		p.launchProvider(configurationChan, pool, p.fileProvider)
	}

	for _, prd := range p.providers {
		safe.Go(func() {
			p.launchProvider(configurationChan, pool, prd)
		})
	}

	return nil
}

func (p *ProviderAggregator) GetConfig() (dynamic.Message, error) {
	var dyConfig = new(dynamic.Configuration)
	//合并所有动态配置文件
	if p.fileProvider != nil {
		dyConfig = resolveProviderConfig(p.fileProvider, dyConfig)
	}
	for _, prd := range p.providers {
		dyConfig = resolveProviderConfig(prd, dyConfig)
	}
	return dynamic.Message{
		Configuration: dyConfig,
	}, nil
}

func resolveProviderConfig(c provider.Provider, dyConfig *dynamic.Configuration) *dynamic.Configuration {
	msg, err := c.GetConfig()
	if err != nil {
		log.Log.WithError(err).Error("aggregate GetConfig fail")
		return nil
	}
	dyConfig, err = utils.DeepCopy(dyConfig, msg.Configuration)
	if err != nil {
		log.Log.WithError(err).Error("aggregate DeepCopy fail")
		return nil
	}
	return dyConfig
}

func (p *ProviderAggregator) launchProvider(configurationChan chan<- dynamic.Message, pool *safe.Pool, prd provider.Provider) {
	jsonConf, err := redactor.RemoveCredentials(prd)
	if err != nil {
		log.Log.WithError(err).Debugf("Cannot marshal the provider configuration %T", prd)
	}

	log.Log.Infof("Starting provider %T", prd)
	log.Log.WithFields(map[string]interface{}{
		"config": jsonConf,
	}).Debugf("%T provider configuration", prd)

	if err := maybeThrottledProvide(prd, p.providersThrottleDuration)(configurationChan, pool); err != nil {
		log.Log.WithError(err).Errorf("Cannot start the provider %T", prd)
		return
	}
}
