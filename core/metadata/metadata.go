package metadata

// Service 元数据存储服务封装
type Service interface {
	Close() error
	DB(name string) DB
}

type DB interface {
	Collection(name string) Collection
}

type Collection interface {
}
