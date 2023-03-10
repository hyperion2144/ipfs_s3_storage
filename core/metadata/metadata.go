package metadata

import (
	"context"
)

// Service 元数据存储服务封装
type Service interface {
	Close() error
	DB(name string) DB
}

type DB interface {
	CreateCollection(ctx context.Context, name string) (Collection, error)
	Collection(name string) (Collection, error)
	ListCollection(ctx context.Context) (map[string]Collection, error)
	DeleteCollection(ctx context.Context, name string) error
}

type Collection interface {
}
