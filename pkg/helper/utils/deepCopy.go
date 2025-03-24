package utils

import (
	"github.com/jinzhu/copier"
	"go-faster-gateway/pkg/log"
)

func DeepCopy[T any](dest *T, src *T) (*T, error) {
	if dest == nil {
		dest = new(T)
	}
	err := copier.Copy(dest, src)
	if err != nil {
		log.Log.WithError(err).Error("DeepCopy fail")
		return dest, err
	}
	return dest, err
}

func DeepCopyDefault[T any](dest *T, src *T) *T {
	if dest == nil {
		dest = new(T)
	}
	err := copier.Copy(dest, src)
	if err != nil {
		log.Log.WithError(err).Error("DeepCopyDefault fail")
		return dest
	}
	return dest
}

func DeepCopyMap[T any, R string](dest *T, src map[R]*T) (*T, error) {
	if dest == nil {
		dest = new(T)
	}
	var err error
	for _, configuration := range src {
		err = copier.Copy(dest, configuration)
		if err != nil {
			log.Log.WithError(err).Error("DeepCopyMap fail")
			return dest, err
		}
	}
	return dest, err
}
