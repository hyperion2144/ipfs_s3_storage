package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
)

var _ metadata.DB = (*db)(nil)

type db struct {
	db *mongo.Database
}

func (m *db) Collection(name string) metadata.Collection {
	c := m.db.Collection(name)
	return newCollection(c)
}

func newMongoDB(d *mongo.Database) *db {
	return &db{db: d}
}
