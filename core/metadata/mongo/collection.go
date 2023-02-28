package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
)

var _ metadata.Collection = (*collection)(nil)

type collection struct {
	collection *mongo.Collection
}

func newCollection(c *mongo.Collection) *collection {
	return &collection{
		collection: c,
	}
}
