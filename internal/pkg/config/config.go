package config

import (
	"fmt"
	"github.com/spf13/viper"
	"go-faster-gateway/pkg/helper/env"
	"os"
	filePathExt "path/filepath"
	"strings"
)

func NewConfig(filePath string) *viper.Viper {
	return getConfig(filePath)

}

func getConfig(path string) *viper.Viper {
	path = getConvertConfigFilePath(path)
	conf := viper.New()
	conf.SetConfigFile(path)
	err := conf.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return conf
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
