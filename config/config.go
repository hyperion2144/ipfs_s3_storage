package config

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	// libp2p listen address，example: /ip4/0.0.0.0/tcp/4005
	Address string
	// keypair store path
	Root string
	// bootstrap address
	Bootstrap []string
}

func (c *Config) MultiAddr() multiaddr.Multiaddr {
	listen, _ := multiaddr.NewMultiaddr(c.Address)
	return listen
}

func (c *Config) PrivateNetworkKey() pnet.PSK {
	var path = filepath.Join(c.Root, "swarm.key")

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("[fatal]: open private network secret key failed: %s", err)
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("[fatal]: read private network secret key failed: %s", err)
	}

	return b
}

func (c *Config) PrivateNodeKey() crypto.PrivKey {
	var priv crypto.PrivKey
	var b []byte
	var path = filepath.Join(c.Root, "private.key")

	file, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		priv, _, err = crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			log.Fatalf("[fatal]: generate private key failed: %s", err)
		}
		b, err = priv.Raw()
		if err != nil {
			log.Fatalf("[fatal]: read private key failed: %s", err)
		}
		err = os.WriteFile(path, b, os.ModePerm)
		if err != nil {
			log.Fatalf("[fatal]: save private key file failed: %s", err)
		}
	} else if err == nil {
		defer file.Close()

		b, err = io.ReadAll(file)
		if err != nil {
			log.Fatalf("[fatal]: read private key file failed: %s", err)
		}
		priv, err = crypto.UnmarshalRsaPrivateKey(b)
		if err != nil {
			log.Fatalf("[fatal]: load private key failed: %s", err)
		}
	}
	if err != nil {
		log.Fatalf("[fatal]: open private key failed: %s", err)
	}

	return priv
}

func (c *Config) BootstrapMultiAddrs() []peer.AddrInfo {
	var bootstrapPeers []multiaddr.Multiaddr
	for _, s := range c.Bootstrap {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			log.Fatalf("[fatal]: load bootstrap address failed: %s", err)
		}
		bootstrapPeers = append(bootstrapPeers, ma)
	}
	peers, _ := peer.AddrInfosFromP2pAddrs(bootstrapPeers...)
	return peers
}
