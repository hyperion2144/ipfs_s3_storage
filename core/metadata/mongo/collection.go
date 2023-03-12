package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
)

var _ metadata.Collection = (*collection)(nil)

type collection struct {
	collection *mongo.Collection
}

func (c *collection) Name() string {
	return c.collection.Name()
}

func (c *collection) Options(ctx context.Context) (metadata.CollectionOptions, error) {
	result := c.collection.FindOne(ctx, bson.M{
		"id": metadata.CollectionOptionID,
	})

	opts := metadata.CollectionOptions{}
	err := result.Decode(&opts)

	return opts, err
}

func newCollection(c *mongo.Collection) *collection {
	return &collection{
		collection: c,
	}
}
