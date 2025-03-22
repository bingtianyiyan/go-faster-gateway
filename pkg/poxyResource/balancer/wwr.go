package balancer

import (
	"math"
	"math/rand"
	"sort"
)

func init() {
	factories[WWRBalancer] = NewWWR
}

type WWR struct {
	BaseBalancer
}

func NewWWR(hosts []*Node) Balancer {
	return &IPHash{
		BaseBalancer: BaseBalancer{
			hosts: hosts,
		},
	}
}

type Chooser struct {
	data   []*Node
	totals []int
	max    int
}

type nodes []*Node

func (a nodes) Len() int {
	return len(a)
}

func (a nodes) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a nodes) Less(i, j int) bool {
	return a[i].Weight < a[j].Weight
}

// Balance selects a suitable host according
func (r *WWR) Balance(_ string) (*Node, error) {
	r.RLock()
	defer r.RUnlock()
	if len(r.hosts) == 0 {
		return nil, NoHostError
	}
	var result []*Node
	mw := 0
	for _, host := range r.hosts {
		if host.Healthy && host.Weight > 0 {
			cw := int(math.Ceil(float64(host.Weight)))
			if cw > mw {
				mw = cw
			}
			result = append(result, host)
		}
	}
	instance := newChooser(result).pick()
	return instance, nil
}

func newChooser(instances []*Node) Chooser {
	sort.Sort(nodes(instances))
	totals := make([]int, len(instances))
	runningTotal := 0
	for i, c := range instances {
		runningTotal += int(c.Weight)
		totals[i] = runningTotal
	}
	return Chooser{data: instances, totals: totals, max: runningTotal}
}

func (chs Chooser) pick() *Node {
	r := rand.Intn(chs.max) + 1
	i := sort.SearchInts(chs.totals, r)
	return chs.data[i]
}
