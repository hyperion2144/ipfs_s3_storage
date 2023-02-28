package mongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
)

var _ metadata.Service = (*Mongo)(nil)

type Mongo struct {
	client *mongo.Client
}

func (m *Mongo) DB(name string) metadata.DB {
	return newMongoDB(m.client.Database(name))
}

func (m *Mongo) Close() error {
	return m.client.Disconnect(context.Background())
}

func NewMongo(address string) *Mongo {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(address))
	if err != nil {
		log.Fatalf("connect to mongo failed: %s", err)
	}

	return &Mongo{client: client}
}
