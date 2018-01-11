package benchers

import (
	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/client/datastor"
	storgrpc "github.com/zero-os/0-stor/client/datastor/grpc"
	"github.com/zero-os/0-stor/client/itsyouonline"
	"github.com/zero-os/0-stor/client/metastor"
	"github.com/zero-os/0-stor/client/metastor/etcd"
	"github.com/zero-os/0-stor/client/metastor/memory"
	"github.com/zero-os/0-stor/client/pipeline"

	log "github.com/Sirupsen/logrus"
)

// newClientFromConfig creates a new zstor client from provided config
// if Metastor shards are empty, it will use an in memory metadata server
func newClientFromConfig(cfg *client.Config, jobCount int, enableCaching bool) (*client.Client, error) {
	// create datastor cluster
	datastorCluster, err := createDataClusterFromConfig(cfg, enableCaching)
	if err != nil {
		return nil, err
	}

	// create data pipeline, using our datastor cluster
	dataPipeline, err := pipeline.NewPipeline(cfg.Pipeline, datastorCluster, jobCount)
	if err != nil {
		return nil, err
	}

	// if no metadata shards are given, return an error,
	// as we require a metastor client
	var metastorClient metastor.Client
	if len(cfg.MetaStor.Shards) == 0 {
		log.Info("Using in memory metadata server")
		metastorClient = memory.NewClient()
	} else {
		log.Info("Using etcd metadata server")
		metastorClient, err = etcd.NewClient(cfg.MetaStor.Shards)
		if err != nil {
			return nil, err
		}
	}

	return client.NewClient(metastorClient, dataPipeline), nil
}

func createDataClusterFromConfig(cfg *client.Config, enableCaching bool) (datastor.Cluster, error) {
	if cfg.IYO == (itsyouonline.Config{}) {
		// create datastor cluster without the use of IYO-backed JWT Tokens,
		// this will only work if all shards use zstordb servers that
		// do not require any authentication (run with no-auth flag)
		return storgrpc.NewCluster(cfg.DataStor.Shards, cfg.Namespace, nil)
	}

	// create IYO client
	client, err := itsyouonline.NewClient(cfg.IYO)
	if err != nil {
		return nil, err
	}

	var tokenGetter datastor.JWTTokenGetter
	// create JWT Token Getter (Using the earlier created IYO Client)
	tokenGetter, err = datastor.JWTTokenGetterUsingIYOClient(cfg.IYO.Organization, client)
	if err != nil {
		return nil, err
	}

	if enableCaching {
		// create cached token getter from this getter, using the default bucket size and count
		tokenGetter, err = datastor.CachedJWTTokenGetter(tokenGetter, -1, -1)
		if err != nil {
			return nil, err
		}
	}

	// create datastor cluster, with the use of IYO-backed JWT Tokens
	return storgrpc.NewCluster(cfg.DataStor.Shards, cfg.Namespace, tokenGetter)
}
