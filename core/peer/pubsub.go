package peer

import (
	"context"
	"fmt"
	"os"

	"github.com/golang/protobuf/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const pubsubTopic = "/libp2p/peer/update/1.0.0"

func pubsubHandler(ctx context.Context, sub *pubsub.Subscription) {
	defer sub.Cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			req := &Request{}
			err = proto.Unmarshal(msg.Data, req)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			switch req.Type {
			case Request_SEND_MESSAGE:
			case Request_UPDATE_PEER:
			}
		}
	}
}
