package gateway

import (
	"context"
	"net/http"

	"github.com/hyperion2144/ipfs_s3_storage/config"
	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
	minio "github.com/minio/minio/cmd"
)

type IPFSObjects struct {
	minio.GatewayUnsupported

	ctx context.Context

	node     *Node
	metadata metadata.DB
}

func NewIPFSObjects(
	ctx context.Context,
	address, root string,
	bootstrap []string,
	metadata metadata.DB,
) (*IPFSObjects, error) {
	node, err := NewNode(ctx, &config.Config{
		Address:   address,
		Root:      root,
		Bootstrap: bootstrap,
	})
	if err != nil {
		return nil, err
	}

	return &IPFSObjects{
		ctx:      ctx,
		node:     node,
		metadata: metadata,
	}, nil
}

func (o *IPFSObjects) Shutdown(ctx context.Context) error {
	return nil
}

func (o *IPFSObjects) StorageInfo(ctx context.Context) (minio.StorageInfo, []error) {
	return minio.StorageInfo{}, nil
}

func (o *IPFSObjects) MakeBucketWithLocation(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error {
	b, err := o.metadata.Collection(bucket)

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

func (o *IPFSObjects) GetObjectInfo(ctx context.Context, bucket, object string, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
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
