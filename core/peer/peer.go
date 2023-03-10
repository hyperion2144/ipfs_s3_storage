package peer

import (
	"context"
	"fmt"
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
	discovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/multiformats/go-multiaddr"
	"github.com/op/go-logging"

	"github.com/hyperion2144/ipfs_s3_storage/core/proto"
)

var (
	defaultReprovideInterval = 12 * time.Hour

	logger = logging.MustGetLogger("peer")
)

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	logger.Infof("discovered new peer %s\n", pi.ID.String())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		logger.Infof("error connecting to peer %s: %s\n", pi.ID.String(), err)
	}

	logger.Infof("connected peer %s success", pi.ID.String())
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, "", &discoveryNotifee{h: h})
	return s.Start()
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

	TopicHandler func(*proto.UpdatePeer)
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

	// setup peer discovery
	go p.Discover("")

	// setup local mDNS discovery
	if err := setupDiscovery(host); err != nil {
		logger.Fatal(err)
	}

	err := p.setupPubsub()
	if err != nil {
		logger.Fatalf("setup peer pubsub failed: %s", err)
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

func (p *Peer) RoutedHost() host.Host {
	return p.routedHost
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

// Bootstrap is an optional helper to connect to the given peers and bootstrap21
// the Peer DHT (and Bitswap). This is a best-effort function. Errors are only
// logged and a warning is printed when less than half of the given peers
// could be contacted. It is fine to pass a list where some peers will not be
// reachable.
func (p *Peer) Bootstrap(peers []peer.AddrInfo) {
	err := p.dht.Bootstrap(p.ctx)
	if err != nil {
		logger.Error(err)
		return
	}

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
}

func (p *Peer) Discover(rendezvous string) {
	var routingDiscovery = discovery.NewRoutingDiscovery(p.dht)

	util.Advertise(p.ctx, routingDiscovery, rendezvous)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:

			peers, err := util.FindPeers(p.ctx, routingDiscovery, rendezvous)
			if err != nil {
				logger.Error(err)
				continue
			}

			for _, pi := range peers {
				if pi.ID == p.host.ID() {
					continue
				}
				if p.host.Network().Connectedness(pi.ID) != coreNetwork.Connected {
					_, err = p.host.Network().DialPeer(p.ctx, pi.ID)
					logger.Infof("Connected to peer %s\n", pi.ID.String())

					if err != nil {
						logger.Errorf("Connect peer %s failed: %s\n", pi.ID.String(), err)
						continue
					}

					logger.Infof("Connected peer %s success.", pi.ID.String())
				}
			}
		}
	}
}
