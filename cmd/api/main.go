package main

import (
	"os"

	_ "github.com/hyperion2144/ipfs_s3_storage/gateway"

	minio "github.com/minio/minio/cmd"
)

func main() {
	minio.Main(os.Args)
}
