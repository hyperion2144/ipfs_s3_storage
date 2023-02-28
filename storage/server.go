package storage

import (
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/pnet"

	"github.com/hyperion2144/ipfs_s3_storage/core/peer"
)

type Config struct {
	Address string
	Secret  pnet.PSK
}

type Node struct {
	p *peer.Peer

	host host.Host
}
