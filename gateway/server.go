package gateway

import (
	"context"

	"github.com/hyperion2144/ipfs_s3_storage/core/metadata"
	"github.com/hyperion2144/ipfs_s3_storage/core/metadata/mongo"
	"github.com/minio/cli"
	"github.com/minio/madmin-go"
	minio "github.com/minio/minio/cmd"
)

const IPFSBackendGateway = "Gateway-GATEWAY"

func init() {
	const ipfsGatewayTemplate = `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} {{if .VisibleFlags}}[FLAGS]{{end}}
{{if .VisibleFlags}}
FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}

EXAMPLES:
  1. Start minio gateway server for Gateway backend
     {{.Prompt}} {{.EnvVarSetCommand}} MINIO_ROOT_USER{{.AssignmentOperator}}accesskey
     {{.Prompt}} {{.EnvVarSetCommand}} MINIO_ROOT_PASSWORD{{.AssignmentOperator}}secretkey
`

	_ = minio.RegisterGatewayCommand(cli.Command{
		Name:               IPFSBackendGateway,
		Usage:              "Network-attached storage (Gateway)",
		Action:             ipfsGatewayMain,
		CustomHelpTemplate: ipfsGatewayTemplate,
		HideHelpCommand:    true,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "path",
				Usage: "root path",
			},
			cli.StringSliceFlag{
				Name:  "bootstrap",
				Usage: "bootstrap peers address",
			},
			cli.StringFlag{
				Name:  "mongo",
				Usage: "mongo address",
			},
		},
	})
}

// Handler for 'minio gateway nas' command line.
func ipfsGatewayMain(ctx *cli.Context) {
	// Validate gateway arguments.
	if !ctx.Args().Present() || ctx.Args().First() == "help" {
		cli.ShowCommandHelpAndExit(ctx, IPFSBackendGateway, 1)
	}

	minio.StartGateway(ctx, &Gateway{
		address:   ctx.GlobalString("address"),
		root:      ctx.String("path"),
		bootstrap: ctx.StringSlice("bootstrap"),
		mongo:     ctx.String("mongo"),
	})
}

// Gateway implements Gateway.
type Gateway struct {
	address   string
	root      string
	bootstrap []string
	mongo     string
}

// Name implements Gateway interface.
func (g *Gateway) Name() string {
	return IPFSBackendGateway
}

// NewGatewayLayer returns nas gatewaylayer.
func (g *Gateway) NewGatewayLayer(madmin.Credentials) (minio.ObjectLayer, error) {
	newObject, err := NewIPFSObjects(minio.GlobalContext, g.address, g.root, g.bootstrap, mongo.NewMongo(g.mongo).DB("metadata"))
	if err != nil {
		return nil, err
	}
	return &ipfsObjects{newObject}, nil
}

// IsListenSupported returns whether listen bucket notification is applicable for this gateway.
func (n *ipfsObjects) IsListenSupported() bool {
	return false
}

func (n *ipfsObjects) StorageInfo(ctx context.Context) (si minio.StorageInfo, _ []error) {
	si, errs := n.ObjectLayer.StorageInfo(ctx)
	si.Backend.GatewayOnline = si.Backend.Type == madmin.Gateway
	si.Backend.Type = madmin.Gateway
	return si, errs
}

// ipfsObjects implements gateway for MinIO and S3 compatible object storage servers.
type ipfsObjects struct {
	minio.ObjectLayer
}

func (n *ipfsObjects) IsTaggingSupported() bool {
	return true
}
