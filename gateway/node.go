package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"github.com/op/go-logging"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/datastore"
	"github.com/hyperion2144/ipfs_s3_storage/core/peer"
	"github.com/hyperion2144/ipfs_s3_storage/core/proto"
)

const (
	deadline = time.Minute
)

var logger = logging.MustGetLogger("gateway/node")

type NodeManager struct {
	nodes map[string]time.Time
	sync.RWMutex
}

func (m *NodeManager) FindPeerHandle(p *proto.UpdatePeer) {
	m.Lock()
	{
		m.nodes[p.GetCid()] = time.Now()
	}
	m.Unlock()
}

func (m *NodeManager) DeadlineTask() {
	m.Lock()
	{
		now := time.Now()
		for cid, timeline := range m.nodes {
			if timeline.Add(deadline).Before(now) {
				logger.Errorf("peer %s is expired", cid)

				delete(m.nodes, cid)
			}
		}
	}
	m.Unlock()
}

// Node gateway Node.
type Node struct {
	ctx context.Context

	peer    *peer.Peer
	manager *NodeManager

	host host.Host
}

func NewNode(ctx context.Context, cfg *config.Config) (*Node, error) {
	// setup a libp2p host, and routedhost peer.
	ds := datastore.NewInMemoryDatastore()
	h, dht, err := peer.SetupLibp2p(
		ctx,
		cfg.PrivateNodeKey(),
		cfg.PrivateNetworkKey(),
		[]multiaddr.Multiaddr{
			cfg.MultiAddr(),
		},
		ds,
		peer.Libp2pOptionsExtra...,
	)
	if err != nil {
		return nil, err
	}

	manager := &NodeManager{
		nodes: make(map[string]time.Time),
	}

	p := peer.New(ctx, ds, h, dht, &peer.Config{
		Offline:           false,
		ReprovideInterval: 12 * time.Hour,
		TopicHandler:      manager.FindPeerHandle,
	})
	p.Bootstrap(cfg.BootstrapMultiAddrs())

	return &Node{
		ctx:     ctx,
		peer:    p,
		manager: manager,
		host:    h,
	}, nil
}

func (n *Node) Run() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			n.manager.DeadlineTask()
		}
	}
}

func (n *Node) Nodes() []string {
	var nodes []string

	n.manager.RLock()
	{
		for cid := range n.manager.nodes {
			nodes = append(nodes, cid)
		}
	}
	n.manager.RUnlock()

	return nodes
}
