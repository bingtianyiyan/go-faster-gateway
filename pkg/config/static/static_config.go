package static

import (
	"fmt"
	"go-faster-gateway/pkg/helper/env"
	"go-faster-gateway/pkg/helper/parser"
	"go-faster-gateway/pkg/helper/utils"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider/file"
	"os"
	filePathExt "path/filepath"
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

func NewStaticConfig(filePath string) (*Configuration, error) {
	c := &Configuration{}
	filePath = getConvertConfigFilePath(filePath)
	err := utils.GetFile(filePath, c)
	return c, err
}

// getConvertConfigFilePath 处理转化文件名
func getConvertConfigFilePath(filePath string) string {
	//环境变量判断
	envDefaultName := env.ModeProd.String()
	envName := os.Getenv("EnvName")
	//如果环境变量为空则返回默认配置文件地址
	if envName == "" {
		return filePath
	}

	envDefaultName = envName

	sep := "/"
	part1 := ""
	// 找到最后一个分隔符的位置
	index := strings.LastIndex(filePath, sep)
	if index >= 0 {
		// 使用最后一个分隔符位置对字符串进行分割
		part1 = filePath[:index]
		fmt.Println(part1)
	}
	//获取文件格式
	var fileExt = filePathExt.Ext(filePath)
	if part1 != "" {
		filePath = part1 + "/" + "settings." + envDefaultName + fileExt
	} else {
		filePath = "settings." + envDefaultName + fileExt
	}
	return filePath
}
