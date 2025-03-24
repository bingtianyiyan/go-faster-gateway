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

func GetViperFile(filename string, element interface{}) (*viper.Viper, error) {
	var err error
	v := viper.New()
	v.AddConfigPath(path.Dir(filename))
	v.SetConfigName(path.Base(filename))
	ext := path.Ext(filename)
	v.SetConfigType(strings.TrimLeft(ext, "."))
	if err = v.ReadInConfig(); err != nil {
		return v, err
	}
	if err = v.Unmarshal(element); err != nil {
		return v, err
	}
	return v, err
}
