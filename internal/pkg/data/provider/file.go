package provider

import (
	"context"
	"errors"
	"fmt"
	"go-faster-gateway/internal/pkg/data"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/poxyResource/balancer"
)

var _ data.IRouteResourceData = (*RouteResourceFileData)(nil)

// RouteResourceFileData 静态文件获取路由数据
type RouteResourceFileData struct {
	conf dynamic.Configuration
}

func NewRouteResourceFileData(conf dynamic.Configuration) data.IRouteResourceData {
	return &RouteResourceFileData{
		conf: conf,
	}
}

func (api *RouteResourceFileData) GetAllList(ctx context.Context) ([]*dynamic.ServiceRoute, error) {
	if api.conf.EasyServiceRoute.Services == nil {
		return nil, errors.New("file with route data is empty")
	}
	var list = make([]*dynamic.ServiceRoute, 0)
	for k, v := range api.conf.EasyServiceRoute.Services {
		v.RouteName = fmt.Sprintf("%s_%s", k, v.RouteName)
		if len(v.BalanceMode) == 0 {
			if len(api.conf.BalanceMode) == 0 {
				v.BalanceMode = balancer.WWRBalancer
			} else {
				v.BalanceMode = api.conf.BalanceMode
			}
		}
		list = append(list, v)
	}
	return list, nil
}
