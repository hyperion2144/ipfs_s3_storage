package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/op/go-logging"
	"google.golang.org/grpc"

	gostream "github.com/libp2p/go-libp2p-gostream"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/proto"
	"github.com/hyperion2144/ipfs_s3_storage/core/protocol"
	"github.com/hyperion2144/ipfs_s3_storage/storage"
)

var logger = logging.MustGetLogger("storage/main")

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	address := flag.String("address", os.Getenv("NODE_ADDRESS"), "node listen address")
	root := flag.String("root", os.Getenv("NODE_ROOT_PATH"), "node root file path")
	ipfsAddress := flag.String("ipfs", "/ip4/127.0.0.1/tcp/5001", "node root file path")
	bootstrap := flag.String("boostrap", "", "bootstrap peer address list")

	flag.Parse()

	if *address == "" {
		*address = "/ip4/0.0.0.0/tcp/5090"
	}
	if *root == "" {
		*root, _ = os.UserHomeDir()
		*root = filepath.Join(*root, ".storage-node")
	}

	var bootstraps []string
	if *bootstrap != "" {
		bootstraps = strings.Split(*bootstrap, ";")
	}

	ipfs := protocol.NewIPFS(*ipfsAddress)
	node, err := storage.NewNode(ctx, ipfs, &config.Config{
		Address:   *address,
		Root:      *root,
		Bootstrap: bootstraps,
	})
	if err != nil {
		logger.Fatal(err)
	}

	listener, err := gostream.Listen(node.Peer.RoutedHost(), proto.FileChannelProtocol)
	if err != nil {
		logger.Fatal(err)
	}

	go node.Background()

	rpc := grpc.NewServer()
	proto.RegisterFileChannelServer(rpc, node)
	if err := rpc.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		logger.Fatal(err)
	}
}
