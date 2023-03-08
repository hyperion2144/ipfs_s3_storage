package peer

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	coreNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	"github.com/op/go-logging"
)

var (
	defaultReprovideInterval = 12 * time.Hour

	logger = logging.MustGetLogger("peer")
)

type mdnsNotifee struct {
	h   host.Host
	ctx context.Context
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	logger.Infof("find peer %s, connecting...", pi.String())

	err := m.h.Connect(m.ctx, pi)
	if err != nil {
		logger.Errorf("connect to %s failed: %v", pi.String(), err)
	}
}

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

	topic      *pubsub.Topic
	routedHost *rhost.RoutedHost
}

func New(
	ctx context.Context,
	datastore datastore.Batching,
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
	logger.Info("I can be reached at:")
	for _, addr := range addrs {
		logger.Info(addr.Encapsulate(hostAddr))
	}

	// Set a stream handler on host A. MessageProtocol is
	// a user-defined protocol name.
	p.routedHost.SetStreamHandler(MessageProtocol, messageHandler(p.cfg.MessageHandler))
	p.routedHost.SetStreamHandler(StreamProtocol, streamHandler(p.cfg.StreamHandler))

	err := p.setupPubsub()
	if err != nil {
		logger.Fatalf("setup peer pubsub failed: %s", err)
	}

	mdnsService := mdns.NewMdnsService(host, "", &mdnsNotifee{h: host, ctx: ctx})
	if err = mdnsService.Start(); err != nil {
		logger.Fatalf("start mdns service failed: %s", err)
	}

	go p.autoclose()

	return p
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

	logger.Info("opening stream")

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
				logger.Warning(err)
				return
			}
			logger.Infof("Connected to %s", pinfo.ID)
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
		logger.Warningf("only connected to %d bootstrap peers out of %d", i, nPeers)
	}

	err := p.dht.Bootstrap(p.ctx)
	if err != nil {
		logger.Error(err)
		return
	}
}
