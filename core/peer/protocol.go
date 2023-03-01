package peer

import (
	"bufio"
	"fmt"
	"io"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

const (
	MessageProtocol = "/libp2p/peer/message/1.0.0"
	StreamProtocol  = "/libp2p/peer/stream/1.0.0"
)

func messageHandler(handler func(message *SendMessage) error) func(s network.Stream) {
	return func(s network.Stream) {
		if handler == nil {
			return
		}
		defer s.Close()

		data, _ := io.ReadAll(bufio.NewReader(s))
		req := &SendMessage{}
		err := proto.Unmarshal(data, req)
		if err != nil {
			m, _ := proto.Marshal(&MessageResponse{
				Err: &Error{Detail: fmt.Sprintf("bad request: %s", err)},
			})
			_, _ = s.Write(m)
			return
		}

		if err = handler(req); err != nil {
			m, _ := proto.Marshal(&MessageResponse{
				Err: &Error{Detail: err.Error()},
			})
			_, _ = s.Write(m)
		}
	}
}

func streamHandler(handler func(reader io.Reader) (string, error)) func(s network.Stream) {
	return func(s network.Stream) {
		if handler == nil {
			return
		}
		log.Println("Got a new stream!")
		defer s.Close()

		if cid, err := handler(bufio.NewReader(s)); err != nil {
			m, _ := proto.Marshal(&StreamResponse{
				Result: &StreamResponse_Err{
					Err: &Error{Detail: err.Error()},
				},
			})
			_, _ = s.Write(m)
		} else {
			m, _ := proto.Marshal(&StreamResponse{
				Result: &StreamResponse_Cid{
					Cid: cid,
				},
			})
			_, _ = s.Write(m)
		}
	}
}
