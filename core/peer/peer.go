package peer

import (
	"context"
	"log"
	"time"

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	exchange "github.com/ipfs/go-ipfs-exchange-interface"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	provider "github.com/ipfs/go-ipfs-provider"
	"github.com/ipfs/go-ipfs-provider/queue"
	"github.com/ipfs/go-ipfs-provider/simple"
	"github.com/ipfs/go-libipfs/bitswap"
	"github.com/ipfs/go-libipfs/bitswap/network"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
)

var (
	defaultReprovideInterval = 12 * time.Hour
)

// Config wraps configuration options for the Peer.
type Config struct {
	Offline bool
	// ReprovideInterval sets how often to reprovide records to the DHT
	ReprovideInterval time.Duration
	// Disables wrapping the blockstore in an ARC cache + Bloomfilter. Use
	// when the given blockstore or datastore already has caching, or when
	// caching is not needed.
	UncachedBlockstore bool
}

func (cfg *Config) setDefaults() {
	if cfg.ReprovideInterval == 0 {
		cfg.ReprovideInterval = defaultReprovideInterval
	}
}

type Peer struct {
	ctx context.Context

	cfg *Config

	host  host.Host
	dht   routing.Routing
	store datastore.Batching

	bstore     blockstore.Blockstore
	bserv      blockservice.BlockService
	exch       exchange.Interface
	reprovider provider.System

	topic      *pubsub.Topic
	routedHost *rhost.RoutedHost
}

func NewPeer(
	ctx context.Context,
	datastore datastore.Batching,
	blockstore blockstore.Blockstore,
	host host.Host,
	dht routing.Routing,
	cfg *Config,
) *Peer {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg.setDefaults()

	p := &Peer{
		ctx:        ctx,
		store:      datastore,
		cfg:        cfg,
		host:       host,
		dht:        dht,
		routedHost: rhost.Wrap(host, dht),
	}

	// Set a stream handler on host A. senderProtocol is
	// a user-defined protocol name.
	p.routedHost.SetStreamHandler(senderProtocol, streamHandler)

	err := p.setupPubsub()
	if err != nil {
		log.Fatalf("setup peer pubsub failed: %s", err)
	}

	err = p.setupBlockstore(blockstore)
	if err != nil {
		log.Fatalf("setup peer blockstore failed: %s", err)
	}

	err = p.setupReprovider()
	if err != nil {
		log.Fatalf("setup peer reprovider failed: %s", err)
	}

	go p.autoclose()

	return p
}

func (p *Peer) setupBlockstore(bs blockstore.Blockstore) error {
	var err error
	if bs == nil {
		bs = blockstore.NewBlockstore(p.store)
	}

	// Support Identity multihashes.
	bs = blockstore.NewIdStore(bs)

	if !p.cfg.UncachedBlockstore {
		bs, err = blockstore.CachedBlockstore(p.ctx, bs, blockstore.DefaultCacheOpts())
		if err != nil {
			return err
		}
	}
	p.bstore = bs
	return nil
}

func (p *Peer) setupBlockService() error {
	if p.cfg.Offline {
		p.bserv = blockservice.New(p.bstore, offline.Exchange(p.bstore))
		return nil
	}

	bswapnet := network.NewFromIpfsHost(p.host, p.dht)
	bswap := bitswap.New(p.ctx, bswapnet, p.bstore)
	p.bserv = blockservice.New(p.bstore, bswap)
	p.exch = bswap
	return nil
}

func (p *Peer) setupReprovider() error {
	if p.cfg.Offline || p.cfg.ReprovideInterval < 0 {
		p.reprovider = provider.NewOfflineProvider()
		return nil
	}

	newQueue, err := queue.NewQueue(p.ctx, "repro", p.store)
	if err != nil {
		return err
	}

	prov := simple.NewProvider(
		p.ctx,
		newQueue,
		p.dht,
	)

	reprov := simple.NewReprovider(
		p.ctx,
		p.cfg.ReprovideInterval,
		p.dht,
		simple.NewBlockstoreProvider(nil),
	)

	p.reprovider = provider.NewSystem(prov, reprov)
	p.reprovider.Run()
	return nil
}

func (p *Peer) setupPubsub() error {
	ps, err := pubsub.NewGossipSub(p.ctx, p.host)
	if err != nil {
		return err
	}
	p.topic, err = ps.Join(pubsubTopic)
	if err != nil {
		return err
	}
	sub, err := p.topic.Subscribe()
	if err != nil {
		return err
	}
	go pubsubHandler(p.ctx, sub)

	return nil
}

func (p *Peer) autoclose() {
	<-p.ctx.Done()
	_ = p.reprovider.Close()
	_ = p.bserv.Close()
	_ = p.topic.Close()
}
