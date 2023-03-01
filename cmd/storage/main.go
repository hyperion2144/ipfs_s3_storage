package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/protocol"
	"github.com/hyperion2144/ipfs_s3_storage/storage"
)

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
		*root = filepath.Join(*root, ".s3node")
	}

	ipfs := protocol.NewIPFS(*ipfsAddress)
	node, err := storage.NewNode(ctx, ipfs, &config.Config{
		Address:   *address,
		Root:      *root,
		Bootstrap: strings.Split(*bootstrap, ";"),
	})
	if err != nil {
		panic(err)
	}

	for {
		err = node.UpdatePeer()
		if err != nil {
			log.Printf("[warn]: update peer info failed: %s", err)
		}
		time.Sleep(time.Second * 10)
	}
}
