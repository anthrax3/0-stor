# 0-stor

[![GoDoc](https://godoc.org/github.com/zero-os/0-stor?status.svg)](https://godoc.org/github.com/zero-os/0-stor) [![Build Status](https://travis-ci.org/zero-os/0-stor.png?branch=master)](https://travis-ci.org/zero-os/0-stor) [![codecov](https://codecov.io/gh/zero-os/0-stor/branch/master/graph/badge.svg)](https://codecov.io/gh/zero-os/0-stor) [![Sourcegraph](https://sourcegraph.com/github.com/zero-os/0-stor/-/badge.svg)](https://sourcegraph.com/github.com/zero-os/0-stor?badge)

A Single device object store.

[link to group on telegram](https://t.me/joinchat/BwOvOw2-K4AN7p9VZckpFw)

## Components

## Server

The 0-stor server is a generic object store that provide simple storage primitives, read, write, list, delete.

0-stor uses [badger](https://github.com/dgraph-io/badger) as the backend key value store. Badger allows storing the keys and the value onto separate devices. Because of this separation, the LSM (Log-Structured Merge) tree of keys can most of the time stay in memory. Typically the keys could be kept in memory and depending on the use case, the values could be served from an SSD or HDD.

### Installation
Install the 0-stor server
```
go get -u github.com/zero-os/0-stor/cmd/zerostorserver
```

### How to run the server

## Running the server

Here are the options of the server:
```
   --debug, -d               Enable debug logging
   --bind value, -b value    Bind address (default: ":8080")
   --data value              Data directory (default: ".db/data")
   --meta value              Metadata directory (default: ".db/meta")
   --profile-addr value      Enables profiling of this server as an http service
   --auth-disable            Disable JWT authentification [$STOR_TESTING]
   --max-msg-size value      Configure the maximum size of the message GRPC server can receive, in MiB (default: 32)
   --async-write             enable asynchronous writes (default: false)
   --help, -h                show help
   --version, -v             print the version

```

Start the server with listening on all interfaces and port 12345
```shell
./zerostorserver --bind :12345 --data /path/to/data --meta /path/to/meta
```

## Client

The client contains all the logic to communicate with the 0-stor server.

The client provides some basic storage primitives to process your data before sending it to the 0-stor server:
- chunking
- compression
- encryption
- replication
- distribution/erasure coding

All of these primitives are configurable and you can decide how your data will be processed before being sent to the 0-stor.

### etcd

Other then a 0-stor server cluster, 0-stor clients also needs an [etcd](https://github.com/coreos/etcd) server cluster running to store it's metadata onto.

To install and run an etcd cluster, check out the [etcd documentation](https://github.com/coreos/etcd#getting-etcd).

### Client API

Client API documentation can be found in the godocs:

[![godoc](https://godoc.org/github.com/zero-os/0-stor/client?status.svg)](https://godoc.org/github.com/zero-os/0-stor/client)

### Client CLI

You can find a CLI for the client in `cmd/zerostorcli`.

To install
```
go get -u github.com/zero-os/0-stor/cmd/zerostorcli
```

### More documentation

You can find more information about the different components in the `/docs` folder of this repository:

* [Server docs](docs/README.md)
* [Client docs](client/README.md)
* [CLI docs](cmd/zerostorcli/README.md)
