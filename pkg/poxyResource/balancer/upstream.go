package balancer

import (
	"go-faster-gateway/internal/pkg/ecode"
	"strconv"
	"sync"
)

// 上游负载均衡器，服务发现时会将服务缓存至upstreams map中，当进行负载均衡时，按照一定的规则进行负载均衡
type Upstream struct {
	LB        map[string]Balancer
	SyncNodes SyncNodesCh
	mu        sync.RWMutex
}

type SyncNodesCh chan NodeServer

type NodeServer struct {
	ServiceName string
	Nodes       []*Node
}

// 一个后端结点对应一个Upstream
type Node struct {
	Service string // 服务
	Port    uint32 // 端口
	Weight  int32  // 权重
	Healthy bool   // 是否健康
}

func (u *Upstream) Watcher(ch SyncNodesCh) {
	for {
		select {
		case data := <-ch:
			u.mu.Lock()
			for _, v := range data.Nodes {
				u.LB[v.Service].Add(v)
			}
			u.mu.Lock()
		default:

		}
	}
}

func (u *Upstream) AddToLB(service string, nodes []*Node, algorithm string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, ok := u.LB[service]
	if !ok {
		u.LB[service], _ = Build(algorithm, nodes)
	}
	for _, v := range nodes {
		u.LB[service].Add(v)
	}
}

func (u *Upstream) GetNextUpstream(service string) (string, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if lb, ok := u.LB[service]; ok {
		upstreamServer, err := lb.Balance(service)
		if err != nil {
			return "", err
		}
		return upstreamServer.Service + ":" + strconv.Itoa(int(upstreamServer.Port)), nil
	}
	return "", ecode.UpstreamNotInit
}
