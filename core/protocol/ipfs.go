package protocol

import (
	"context"
	"io"

	httpclient "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/multiformats/go-multiaddr"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("protocol/ipfs")

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

func (protocol *IPFS) Del(cid string) error {
	return protocol.client.Pin().Rm(context.Background(), path.New(cid))
}

func (protocol *IPFS) Move(s, d string) error {
	return protocol.client.Pin().Update(context.Background(), path.New(s), path.New(d))
}

func NewIPFS(addr string) *IPFS {
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		logger.Fatalf("parse ipfs multiaddr failed: %s", err)
	}
	api, err := httpclient.NewApi(maddr)
	if err != nil {
		logger.Fatalf("create ipfs http client failed: %s", err)
	}
	return &IPFS{client: api}
}
