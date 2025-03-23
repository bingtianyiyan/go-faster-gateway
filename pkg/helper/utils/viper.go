package utils

import (
	"github.com/spf13/viper"
	"path"
	"strings"
)

func GetFile(filename string, element interface{}) error {
	var err error
	viper.AddConfigPath(path.Dir(filename))
	viper.SetConfigName(path.Base(filename))
	ext := path.Ext(filename)
	viper.SetConfigType(strings.TrimLeft(ext, "."))

	if err = viper.ReadInConfig(); err != nil {
		return err
	}
	if err = viper.Unmarshal(element); err != nil {
		return err
	}
	return err
}
