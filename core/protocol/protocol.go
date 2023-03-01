package protocol

import (
	"io"
)

type CidInfo struct {
	Bytes string `json:"Bytes"`
	Hash  string `json:"Hash"`
	Name  string `json:"Name"`
	Size  string `json:"Size"`
}

// Protocol 存储协议封装
type Protocol interface {
	Add(reader io.Reader) (string, error)
	Cat(cid string) (io.ReadCloser, error)
	Del(cid string) error
	Move(s, d string) error
}
