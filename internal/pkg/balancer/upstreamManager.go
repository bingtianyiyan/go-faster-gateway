package balancer

import (
	"errors"
	"go-faster-gateway/internal/pkg/ecode"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/poxyResource/balancer"
)

// UpstreamManager
type UpstreamManager struct {
	Upstreams *balancer.Upstream // 上游服务，一般路由会保存上游服务的名称，转发到对应的上游服务上去，可以使用负载均衡算法
}

func NewUpstreamManager() *UpstreamManager {
	return &UpstreamManager{
		Upstreams: &balancer.Upstream{
			LB:        make(map[string]balancer.Balancer),
			SyncNodes: make(chan balancer.NodeServer, 1),
		},
	}
}

// GetLBUpstream 获取负载均衡后的上游服务
func (f *UpstreamManager) GetLBUpstream(serviceName string, config *dynamic.Configuration) (string, error) {
	var (
		us  string
		err error
	)
	us, err = f.Upstreams.GetNextUpstream(serviceName)
	if err != nil && !errors.Is(err, ecode.UpstreamNotInit) {
		return "", err
	}
	var serviceMap = config.HTTP.Services[serviceName]
	var modelNode = func(model *dynamic.Service) []*balancer.Node {
		nodes := make([]*balancer.Node, 0)
		for _, v := range model.Servers {
			node := &balancer.Node{
				Service: v.Host,
				Port:    uint32(v.Port),
				Weight:  int32(v.Weight),
				Healthy: v.Healthy,
			}
			nodes = append(nodes, node)
		}
		return nodes
	}
	nodes := modelNode(serviceMap)
	err = f.Upstreams.AddToLB(serviceName, nodes, config.BalanceMode.Balance)
	if err != nil {
		log.Log.WithError(err).Error("AddToLB fail")
		return us, err
	}
	us, err = f.Upstreams.GetNextUpstream(serviceName)
	return us, err
}

// 获取上游信息
func (f *UpstreamManager) GetUpstream() *balancer.Upstream {
	return f.Upstreams
}
