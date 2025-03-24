package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/config/static"
	"go-faster-gateway/pkg/helper/env"
	"go-faster-gateway/pkg/helper/utils"
	"go-faster-gateway/pkg/log"
	"os"
	filePathExt "path/filepath"
	"strings"
)

// ConfigurationManager 配置管理器
type ConfigurationManager struct {
	filePath        string
	watchStaticFile bool
	staticConfig    *static.Configuration
	dynamicConfig   *dynamic.Configuration
	watch           *ConfigurationWatcher
}

func NewConfigurationManager(filePath string, watchStaticFile bool) *ConfigurationManager {
	return &ConfigurationManager{
		filePath:        filePath,
		watchStaticFile: watchStaticFile,
		//staticConfig:  new(static.Configuration),
		dynamicConfig: new(dynamic.Configuration),
	}
}

func (f *ConfigurationManager) SetWatch(watch *ConfigurationWatcher) {
	f.watch = watch
}

func (f *ConfigurationManager) SetStaticConfig(scConfig *static.Configuration) {
	f.staticConfig = scConfig
}

func (f *ConfigurationManager) SetDynamicConfig(dyConfig *dynamic.Configuration) {
	f.dynamicConfig = dyConfig
}

func (f *ConfigurationManager) GetStaticConfig() *static.Configuration {
	if f.staticConfig != nil {
		return f.staticConfig
	}
	f.newStaticConfig(f.filePath)
	return f.staticConfig
}

func (f *ConfigurationManager) GetDynamicConfig() *dynamic.Configuration {
	return f.dynamicConfig
}

func (f *ConfigurationManager) GetWatcher() *ConfigurationWatcher {
	return f.watch
}

func (f *ConfigurationManager) newStaticConfig(filePath string) error {
	f.staticConfig = new(static.Configuration)
	filePath = getConvertConfigFilePath(filePath)
	v, err := utils.GetViperFile(filePath, f.staticConfig)
	if err == nil && f.watchStaticFile {
		// 设置监听文件变化
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			log.Log.Debugf("Detach static config change:%s", e.Name)
			// 重新解析配置
			if err = v.Unmarshal(f.staticConfig); err != nil {
				log.Log.WithError(err).Errorf("重新解析配置失败: %v", err)
				return
			}
			log.Log.Debugf("new static config: %+v\n", *f.staticConfig)
		})
	} else {
		f.staticConfig = nil
		log.Log.WithError(err).Error("newStaticConfig utils.GetViperFile fail")
	}
	return err
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
