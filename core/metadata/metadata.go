package metadata

import (
	"context"
	"time"
)

const CollectionOptionID = ".options"

type CollectionOptions struct {
	// ID always has to be set as `.options`.
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	DeletedAt  time.Time `json:"deleted_at"`
	Locking    bool      `json:"locking"`
	Versioning bool      `json:"versioning"`
}

// Service 元数据存储服务封装
type Service interface {
	Close() error
	DB(name string) DB
}

type DB interface {
	CreateCollection(ctx context.Context, name string, opts CollectionOptions) (Collection, error)
	Collection(name string) (Collection, error)
	ListCollection(ctx context.Context) (map[string]Collection, error)
	DeleteCollection(ctx context.Context, name string) error
}

type Collection interface {
	Name() string
	Options(ctx context.Context) (CollectionOptions, error)
}
