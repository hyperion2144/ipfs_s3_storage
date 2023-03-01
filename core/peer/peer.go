package peer

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
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
	coreNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
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

	TopicHandler   func(*UpdatePeer)
	MessageHandler func(*SendMessage) error
	StreamHandler  func(reader io.Reader) (string, error)
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

func New(
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

	// Build host multiaddress
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", p.routedHost.ID().String()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	// addr := routedHost.Addrs()[0]
	addrs := p.routedHost.Addrs()
	log.Println("I can be reached at:")
	for _, addr := range addrs {
		log.Println(addr.Encapsulate(hostAddr))
	}

	// Set a stream handler on host A. MessageProtocol is
	// a user-defined protocol name.
	p.routedHost.SetStreamHandler(MessageProtocol, messageHandler(p.cfg.MessageHandler))
	p.routedHost.SetStreamHandler(StreamProtocol, streamHandler(p.cfg.StreamHandler))

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
	go pubsubHandler(p.ctx, sub, p.cfg.TopicHandler)

	return nil
}

func (p *Peer) autoclose() {
	<-p.ctx.Done()
	_ = p.reprovider.Close()
	_ = p.bserv.Close()
	_ = p.topic.Close()
}

func (p *Peer) ID() string {
	return p.routedHost.ID().String()
}

func (p *Peer) HostAddr() string {
	return p.routedHost.ID().String()
}

func (p *Peer) NewStream(proto, target string) (coreNetwork.Stream, error) {
	peerid, err := peer.Decode(target)
	if err != nil {
		return nil, err
	}

	log.Println("opening stream")

	// make a new stream from host B to host
	s, err := p.routedHost.NewStream(
		context.Background(),
		peerid,
		protocol.ConvertFromStrings([]string{proto})...,
	)
	return s, err
}

func (p *Peer) Publish(ctx context.Context, data []byte) error {
	return p.topic.Publish(ctx, data)
}

// Bootstrap is an optional helper to connect to the given peers and bootstrap
// the Peer DHT (and Bitswap). This is a best-effort function. Errors are only
// logged and a warning is printed when less than half of the given peers
// could be contacted. It is fine to pass a list where some peers will not be
// reachable.
func (p *Peer) Bootstrap(peers []peer.AddrInfo) {
	connected := make(chan struct{})

	var wg sync.WaitGroup
	for _, pinfo := range peers {
		//h.Peerstore().AddAddrs(pinfo.ID, pinfo.Addrs, peerstore.PermanentAddrTTL)
		wg.Add(1)
		go func(pinfo peer.AddrInfo) {
			defer wg.Done()
			err := p.host.Connect(p.ctx, pinfo)
			if err != nil {
				log.Printf("[warn]: %s", err)
				return
			}
			log.Printf("[info]: Connected to %s", pinfo.ID)
			connected <- struct{}{}
		}(pinfo)
	}

	go func() {
		wg.Wait()
		close(connected)
	}()

	i := 0
	for range connected {
		i++
	}
	if nPeers := len(peers); i < nPeers/2 {
		log.Printf("[warn]: only connected to %d bootstrap peers out of %d", i, nPeers)
	}

	err := p.dht.Bootstrap(p.ctx)
	if err != nil {
		log.Printf("[error]: %s", err)
		return
	}
}
