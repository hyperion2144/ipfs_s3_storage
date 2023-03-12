package mongo

import (
	"context"

	"github.com/op/go-logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
)

var logger = logging.MustGetLogger("metadata/mongo")

var _ metadata.DB = (*db)(nil)

type db struct {
	db *mongo.Database
}

func (m *db) CreateCollection(
	ctx context.Context,
	name string,
	opts metadata.CollectionOptions,
) (metadata.Collection, error) {
	err := m.db.CreateCollection(ctx, name)
	if err != nil {
		return nil, err
	}

	c := m.db.Collection(name)
	_, err = c.InsertOne(ctx, opts)
	if err != nil {
		return nil, err
	}

	return newCollection(c), nil
}

func (m *db) ListCollection(ctx context.Context) (map[string]metadata.Collection, error) {
	names, err := m.db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	collections := make(map[string]metadata.Collection)
	for _, name := range names {
		c := m.db.Collection(name)
		collections[name] = newCollection(c)
	}
	return collections, nil
}

func (m *db) DeleteCollection(ctx context.Context, name string) error {
	return m.db.Collection(name).Drop(ctx)
}

func (m *db) Collection(name string) (metadata.Collection, error) {
	c := m.db.Collection(name)
	return newCollection(c), nil
}

func newMongoDB(d *mongo.Database) *db {
	return &db{db: d}
}
