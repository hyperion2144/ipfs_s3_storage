package datastore

import (
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
)

// NewInMemoryDatastore provides a sync datastore that lives in-memory only
// and is not persisted.
func NewInMemoryDatastore() datastore.Batching {
	return dssync.MutexWrap(datastore.NewMapDatastore())
}
