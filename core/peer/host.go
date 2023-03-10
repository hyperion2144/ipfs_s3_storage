package peer

import (
	"context"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
)

var connMgr, _ = connmgr.NewConnManager(100, 600, connmgr.WithGracePeriod(time.Minute))

var Libp2pOptionsExtra = []libp2p.Option{
	libp2p.NATPortMap(),
	libp2p.ConnectionManager(connMgr),
	libp2p.EnableNATService(),
}

func SetupLibp2p(
	ctx context.Context,
	hostKey crypto.PrivKey,
	secret pnet.PSK,
	listenAddrs []multiaddr.Multiaddr,
	ds datastore.Batching,
	opts ...libp2p.Option,
) (host.Host, *dht.IpfsDHT, error) {

	var ddht *dht.IpfsDHT
	var err error
	var transports = libp2p.DefaultTransports

	if secret != nil {
		transports = libp2p.ChainOptions(
			libp2p.NoTransports,
			libp2p.Transport(tcp.NewTCPTransport),
			libp2p.Transport(websocket.New),
		)
	}

	finalOpts := []libp2p.Option{
		libp2p.Identity(hostKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.PrivateNetwork(secret),
		transports,
	}
	finalOpts = append(finalOpts, opts...)

	h, err := libp2p.New(
		finalOpts...,
	)
	if err != nil {
		return nil, nil, err
	}

	ddht, err = newDHT(ctx, h, ds)
	if err != nil {
		return nil, nil, err
	}

	return h, ddht, nil
}

func newDHT(ctx context.Context, h host.Host, ds datastore.Batching) (*dht.IpfsDHT, error) {
	dhtOpts := []dht.Option{
		dht.NamespacedValidator("pk", record.PublicKeyValidator{}),
		dht.Concurrency(10),
		dht.Mode(dht.ModeAuto),
		dht.Datastore(ds),
	}

	return dht.New(ctx, h, dhtOpts...)

}
