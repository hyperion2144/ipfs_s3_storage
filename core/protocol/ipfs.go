package protocol

import (
	"context"
	"io"
	"log"

	httpclient "github.com/ipfs/go-ipfs-http-client"
	"github.com/multiformats/go-multiaddr"
)

var _ Protocol = (*IPFS)(nil)

// IPFS ipfs 客户端封装
type IPFS struct {
	client *httpclient.HttpApi
}

func (protocol *IPFS) Add(reader io.Reader) (string, error) {
	var cidInfo CidInfo
	err := protocol.client.Request("add").Body(reader).Exec(context.Background(), &cidInfo)
	return cidInfo.Hash, err
}

func (protocol *IPFS) Cat(cid string) (io.ReadCloser, error) {
	response, err := protocol.client.Request("cat", cid).Send(context.Background())
	if err != nil {
		return nil, err
	}
	return response.Output, response.Error
}

func NewIPFS(addr string) *IPFS {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		log.Fatalf("parse ipfs multiaddr failed: %s", err)
	}
	api, err := httpclient.NewApi(maddr)
	if err != nil {
		log.Fatalf("create ipfs http client failed: %s", err)
	}
	return &IPFS{client: api}
}
