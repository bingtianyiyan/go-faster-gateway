package provider

import (
	"context"
	"errors"
	"fmt"
	"go-faster-gateway/internal/pkg/data"
	"go-faster-gateway/pkg/config/dynamic"
)

var _ data.IRouteResourceData = (*RouteResourceFileData)(nil)

// RouteResourceFileData 静态文件获取路由数据
type RouteResourceFileData struct {
	routeList map[string]*dynamic.ServiceRoute
}

func NewRouteResourceFileData(routes map[string]*dynamic.ServiceRoute) data.IRouteResourceData {
	return &RouteResourceFileData{
		routeList: routes,
	}
}

func (api *RouteResourceFileData) GetAllList(ctx context.Context) ([]*dynamic.ServiceRoute, error) {
	if api.routeList == nil {
		return nil, errors.New("file with route data is empty")
	}
	var list = make([]*dynamic.ServiceRoute, 0)
	for k, v := range api.routeList {
		v.RouteName = fmt.Sprintf("%s_%s", k, v.RouteName)
		list = append(list, v)
	}
	return list, nil
}
