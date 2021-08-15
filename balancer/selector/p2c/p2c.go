package p2c

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/balancer"
)

const (
	forcePick = time.Second * 3
)

var (
	_ balancer.Selector = &Selector{}
)

// statistics is info for log
type statistic struct {
	addr     string
	score    float64
	cs       uint64
	lantency time.Duration
	load     uint64
	inflight int64
	reqs     int64
	predict  time.Duration
}

// New p2c
func New(errHandler func(error) bool) balancer.Selector {
	p := &Selector{
		r:          rand.New(rand.NewSource(time.Now().UnixNano())),
		subConns:   make(map[string]balancer.Node),
		errHandler: errHandler,
	}
	return p
}

// Selector is p2c selector
type Selector struct {
	// subConns is the snapshot of the weighted-roundrobin balancer when this picker was
	// created. The slice is immutable. Each Get() will do a round robin
	// selection from it and return the selected SubConn.
	subConns   map[string]balancer.Node
	logTs      int64
	r          *rand.Rand
	lk         sync.Mutex
	errHandler func(err error) (isErr bool)
}

// choose two distinct nodes
func (s *Selector) prePick(nodes []balancer.Node) (nodeA balancer.Node, nodeB balancer.Node) {
	a := s.r.Intn(len(nodes))
	b := s.r.Intn(len(nodes) - 1)
	if b >= a {
		b = b + 1
	}
	nodeA, nodeB = nodes[a], nodes[b]
	return
}

func (s *Selector) Select(ctx context.Context, nodes []balancer.Node) (balancer.Node, error) {
	var pc, upc balancer.Node
	start := time.Now()

	if len(nodes) == 0 {
		return nil, balancer.ErrNoAvaliable
	} else if len(nodes) == 1 {
		return nodes[0], nil
	} else {
		nodeA, nodeB := s.prePick(nodes)
		// meta.Weight为服务发布者在disocvery中设置的权重
		if nodeB.Weight() > nodeA.Weight() {
			pc, upc = nodeB, nodeA
		} else {
			pc, upc = nodeA, nodeB
		}
		// 如果选中的节点，在forceGap期间内从来没有被选中一次，则强制选一次
		// 利用强制的机会，来触发成功率、延迟的更新
		// TODO: 并发问题导致瞬间多次选择upc
		if start.Sub(upc.LastPick()) > forcePick {
			pc = upc
		}
	}

	return pc, nil
}

/*

func (p *P2cPicker) PrintStats() {
	if len(p.subConns) == 0 {
		return
	}
	stats := make([]statistic, 0, len(p.subConns))
	var serverName string
	var reqs int64
	var now = time.Now().UnixNano()
	for _, conn := range p.subConns {
		var stat statistic
		stat.addr = conn.node.Endpoints[0]
		stat.cs = atomic.LoadUint64(&conn.success)
		stat.inflight = atomic.LoadInt64(&conn.inflight)
		stat.lantency = time.Duration(atomic.LoadInt64(&conn.lag))
		stat.reqs = atomic.SwapInt64(&conn.reqs, 0)
		stat.load = conn.load(now)
		stat.predict = time.Duration(atomic.LoadInt64(&conn.predict))
		stats = append(stats, stat)
		if serverName == "" {
			serverName = conn.node.Name
		}
		reqs += stat.reqs
	}
	if reqs > 10 {
		//log.DefaultLog.Debugf("p2c %s : %+v", serverName, stats)
	}
}
*/
