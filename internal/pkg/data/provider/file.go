package provider

import (
	"context"
	"errors"
	"go-faster-gateway/internal/pkg/data"
	"go-faster-gateway/pkg/config/dynamic"
)

var _ data.IRouteResourceData = (*RouteResourceFileData)(nil)

// RouteResourceFileData 静态文件获取路由数据
type RouteResourceFileData struct {
	routeList map[string]*dynamic.Service
}

func NewRouteResourceFileData(routes map[string]*dynamic.Service) data.IRouteResourceData {
	return &RouteResourceFileData{
		routeList: routes,
	}
}

func (api *RouteResourceFileData) GetAllList(ctx context.Context) ([]*dynamic.Service, error) {
	if len(api.routeList) == 0 {
		return nil, errors.New("file with route data is empty")
	}
	var list = make([]*dynamic.Service, 0)
	for k, v := range api.routeList {
		v.Service = k
		list = append(list, v)
	}
	return list, nil
}
