package static

import (
	"go-faster-gateway/pkg/helper/parser"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider/file"
	"strings"
)

// Configuration is the static configuration.
type Configuration struct {
	//代理地址
	EntryPoint EntryPoint `description:"Entry point definition." json:"entryPoint,omitempty" toml:"entryPoint,omitempty" yaml:"entryPoint,omitempty" export:"true"`
	//其他动态配置文件提供者
	Providers *Providers `description:"Providers configuration." json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty" export:"true"`
	//日志
	Logger *log.Logger `description:"gateway log settings." json:"log,omitempty" toml:"logger,omitempty" yaml:"logger,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Providers contains providers configuration.
type Providers struct {
	//刷新频率
	ProvidersThrottleDuration parser.Duration `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." json:"providersThrottleDuration,omitempty" toml:"providersThrottleDuration,omitempty" yaml:"providersThrottleDuration,omitempty" export:"true"`

	File *file.Provider `description:"Enable File backend with default settings." json:"file,omitempty" toml:"file,omitempty" yaml:"file,omitempty" export:"true"`
}

// ValidateConfiguration validate that configuration is coherent.
func (c *Configuration) ValidateConfiguration() error {
	return nil
}

// EntryPoint holds the entry point configuration.
type EntryPoint struct {
	Address string `description:"Entry point address." json:"address,omitempty" toml:"address,omitempty" yaml:"address,omitempty"`
	Port    int    `description:"Enables EntryPoints from the same or different processes listening on the same TCP/UDP port." json:"port,omitempty" toml:"port,omitempty" yaml:"port,omitempty"`
}

// GetAddress strips any potential protocol part of the address field of the
// entry point, in order to return the actual address.
func (ep *EntryPoint) GetAddress() string {
	splitN := strings.SplitN(ep.Address, "/", 2)
	return splitN[0]
}
