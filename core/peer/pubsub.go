package peer

import (
	"context"

	"github.com/golang/protobuf/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	proto2 "github.com/hyperion2144/ipfs_s3_storage/core/proto"
)

const pubsubTopic = "/libp2p/peer/update/1.0.0"

func pubsubHandler(ctx context.Context, sub *pubsub.Subscription, handler func(*proto2.UpdatePeer)) {
	if handler == nil {
		return
	}
	defer sub.Cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := sub.Next(ctx)
			if err != nil {
				logger.Error(err)
				continue
			}

			req := &proto2.UpdatePeer{}
			err = proto.Unmarshal(msg.Data, req)
			if err != nil {
				logger.Error(err)
				continue
			}

			handler(req)
		}
	}
}
