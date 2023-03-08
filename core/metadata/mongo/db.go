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

func (m *db) CreateCollection(name string) (metadata.Collection, error) {
	err := m.db.CreateCollection(context.Background(), name)
	if err != nil {
		return nil, err
	}

	return m.Collection(name)
}

func (m *db) ListCollection() ([]metadata.Collection, error) {
	names, err := m.db.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	var collections []metadata.Collection
	for _, name := range names {
		collection := m.db.Collection(name)
		collections = append(collections, collection)
	}
	return collections, nil
}

func (m *db) DeleteCollection(name string) error {
	return m.db.Collection(name).Drop(context.Background())
}

func (m *db) Collection(name string) (metadata.Collection, error) {
	c := m.db.Collection(name)
	return newCollection(c), nil
}

func newMongoDB(d *mongo.Database) *db {
	return &db{db: d}
}
