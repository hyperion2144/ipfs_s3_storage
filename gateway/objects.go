package gateway

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	minio "github.com/minio/minio/cmd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
	pb "github.com/hyperion2144/ipfs_s3_storage/core/proto"
)

type IPFSObjects struct {
	minio.GatewayUnsupported

	ctx context.Context

	node     *Node
	metadata metadata.DB

	conns    map[string]*grpc.ClientConn
	connlock sync.RWMutex

	collections map[string]metadata.Collection
	collectlock sync.RWMutex
}

func NewIPFSObjects(
	ctx context.Context,
	address, root string,
	bootstrap []string,
	m metadata.DB,
) (*IPFSObjects, error) {
	node, err := NewNode(ctx, &config.Config{
		Address:   address,
		Root:      root,
		Bootstrap: bootstrap,
	})
	if err != nil {
		return nil, err
	}

	collections, err := m.ListCollection(ctx)
	if err != nil {
		return nil, err
	}

	return &IPFSObjects{
		ctx:         ctx,
		node:        node,
		metadata:    m,
		conns:       make(map[string]*grpc.ClientConn),
		collections: collections,
	}, nil
}

func (o *IPFSObjects) dial(peerID peer.ID) (pb.FileChannelClient, error) {
	o.connlock.Lock()
	defer o.connlock.Unlock()

	if conn, ok := o.conns[peerID.String()]; ok {
		if conn.GetState() == connectivity.Shutdown {
			if err := conn.Close(); err != nil && status.Code(err) != codes.Canceled {
				logger.Errorf("error closing connection: %v", err)
			}
		} else {
			return pb.NewFileChannelClient(conn), nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	conn, err := grpc.DialContext(ctx, peerID.String())
	if err != nil {
		return nil, err
	}
	o.conns[peerID.String()] = conn
	return pb.NewFileChannelClient(conn), nil
}

func (o *IPFSObjects) Shutdown(context.Context) error {
	return nil
}

func (o *IPFSObjects) StorageInfo(context.Context) (minio.StorageInfo, []error) {
	return minio.StorageInfo{}, nil
}

func (o *IPFSObjects) MakeBucketWithLocation(ctx context.Context, bucket string, _ minio.MakeBucketOptions) error {
	b, err := o.metadata.CreateCollection(ctx, bucket)
	if err != nil {
		return err
	}

	o.collectlock.Lock()
	{
		o.collections[bucket] = b
	}
	o.collectlock.Unlock()
	return nil
}

func (o *IPFSObjects) GetBucketInfo(ctx context.Context, bucket string, opts minio.BucketOptions) (bucketInfo minio.BucketInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) ListBuckets(ctx context.Context, opts minio.BucketOptions) (buckets []minio.BucketInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) DeleteBucket(ctx context.Context, bucket string, opts minio.DeleteBucketOptions) error {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) ListObjects(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (result minio.ListObjectsInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) GetObjectNInfo(ctx context.Context, bucket, object string, rs *minio.HTTPRangeSpec, h http.Header, lockType minio.LockType, opts minio.ObjectOptions) (reader *minio.GetObjectReader, err error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) GetObjectInfo(
	ctx context.Context,
	bucket, object string,
	opts minio.ObjectOptions,
) (objInfo minio.ObjectInfo, err error) {

	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) PutObject(ctx context.Context, bucket, object string, data *minio.PutObjReader, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) DeleteObject(ctx context.Context, bucket, object string, opts minio.ObjectOptions) (minio.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) DeleteObjects(ctx context.Context, bucket string, objects []minio.ObjectToDelete, opts minio.ObjectOptions) ([]minio.DeletedObject, []error) {
	//TODO implement me
	panic("implement me")
}

func (o *IPFSObjects) NewMultipartUpload(ctx context.Context, bucket, object string, opts minio.ObjectOptions) (result *minio.NewMultipartUploadResult, err error) {
	//TODO implement me
	panic("implement me")
}
