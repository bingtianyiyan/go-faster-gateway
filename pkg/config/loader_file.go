package config

import (
	"fmt"
	"go-faster-gateway/pkg/helper/env"
	"go-faster-gateway/pkg/helper/file"
	"os"
	filePathExt "path/filepath"
	"strings"
)

// FileLoader loads a configuration from a file.
type FileLoader struct {
	ConfigFileFlag string
	filename       string
}

// GetFilename returns the configuration file if any.
func (f *FileLoader) GetFilename() string {
	return f.filename
}

// Load loads the  configuration from a file from default locations.
func (f *FileLoader) Load(args []string, cmd *Command) (bool, error) {
	configFile, err := loadConfigFiles(args[1], cmd.Configuration)
	if err != nil {
		return false, err
	}

	//处理根据环境变量
	configFile = getConvertConfigFilePath(configFile)
	f.filename = configFile

	if configFile == "" {
		return false, nil
	}

	//log.Log.Infof("Configuration loaded from file: %s", configFile)
	//
	//content, _ := os.ReadFile(configFile)
	//log.Log.Debug("configFile", configFile, "content", string(content))
	fmt.Sprintf("Configuration loaded from file: %s", configFile)
	return true, nil
}

// loadConfigFiles tries to decode  default locations for the configuration file.
func loadConfigFiles(configFile string, element interface{}) (string, error) {
	finder := file.Finder{
		BasePaths:  []string{"$XDG_CONFIG_HOME/go-faster-gateway", "$HOME/.config/go-faster-gateway", "./go-faster-gateway"},
		Extensions: []string{"toml", "yaml", "yml"},
	}

	filePath, err := finder.Find(configFile)
	if err != nil {
		return "", err
	}

	if len(filePath) == 0 {
		return "", nil
	}

	if err := file.Decode(filePath, element); err != nil {
		return "", err
	}
	return filePath, nil
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
