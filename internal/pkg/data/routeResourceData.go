package data

import (
	"context"
	"go-faster-gateway/pkg/config/dynamic"
)

// IRouteResourceData router得数据抽象接口
type IRouteResourceData interface {
	// GetAllList 获取所有服务信息
	GetAllList(ctx context.Context) ([]*dynamic.ServiceRoute, error)
}
