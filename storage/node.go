package storage

import (
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/datastore"
	"github.com/hyperion2144/ipfs_s3_storage/core/peer"
	"github.com/hyperion2144/ipfs_s3_storage/core/protocol"
)

type Node struct {
	p *peer.Peer

	protocol protocol.Protocol

	host host.Host
}

func NewNode(
	ctx context.Context,
	protocol protocol.Protocol,
	cfg *config.Config,
) (*Node, error) {
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

	p := peer.New(ctx, ds, h, dht, &peer.Config{
		ReprovideInterval: 12 * time.Hour,
		MessageHandler: func(message *peer.SendMessage) error {
			switch message.GetType() {
			case peer.SendMessage_Remove:
				return protocol.Del(message.GetCid())
			case peer.SendMessage_Move:
				return protocol.Move(message.GetFrom(), message.GetCid())
			}
			return nil
		},
		StreamHandler: func(reader io.Reader) (string, error) {
			return protocol.Add(reader)
		},
	})
	p.Bootstrap(cfg.BootstrapMultiAddrs())

	return &Node{
		p:        p,
		protocol: protocol,
		host:     h,
	}, nil
}

func (n *Node) Run() error {
	request := peer.UpdatePeer{
		Peer: &peer.UpdatePeer_Peer{
			Cid: n.p.HostAddr(),
		},
	}
	b, err := proto.Marshal(&request)
	if err != nil {
		return err
	}

	return n.p.Publish(context.Background(), b)
}
