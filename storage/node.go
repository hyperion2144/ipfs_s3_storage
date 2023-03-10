package storage

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"github.com/op/go-logging"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/datastore"
	"github.com/hyperion2144/ipfs_s3_storage/core/peer"
	pb "github.com/hyperion2144/ipfs_s3_storage/core/proto"
	"github.com/hyperion2144/ipfs_s3_storage/core/protocol"
)

var logger = logging.MustGetLogger("storage/node")

type Node struct {
	pb.UnimplementedFileChannelServer

	Peer *peer.Peer

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
	})
	p.Bootstrap(cfg.BootstrapMultiAddrs())

	return &Node{
		Peer:     p,
		protocol: protocol,
		host:     h,
	}, nil
}

func (n *Node) Background() error {
	request := pb.UpdatePeer{
		Cid: n.Peer.HostAddr(),
	}
	b, err := proto.Marshal(&request)
	if err != nil {
		return err
	}

	for {
		err := n.Peer.Publish(context.Background(), b)
		if err != nil {
			logger.Warningf("update peer info failed: %s", err)
		}
		time.Sleep(time.Second * 10)
	}

}

func (n *Node) Add(request pb.FileChannel_AddServer) error {
	//TODO implement me
	panic("implement me")
}

func (n *Node) Remove(ctx context.Context, request *pb.RemoveRequest) (*pb.RemoveReply, error) {
	//TODO implement me
	panic("implement me")
}

func (n *Node) Move(ctx context.Context, request *pb.MoveRequest) (*pb.MoveReply, error) {
	//TODO implement me
	panic("implement me")
}
