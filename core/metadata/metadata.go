package metadata

// Service 元数据存储服务封装
type Service interface {
	Close() error
	DB(name string) DB
}

type DB interface {
	CreateCollection(name string) (Collection, error)
	Collection(name string) (Collection, error)
	ListCollection() ([]Collection, error)
	DeleteCollection(name string) error
}

type Collection interface {
}
