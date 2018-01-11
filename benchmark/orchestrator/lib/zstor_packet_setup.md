# Packet deployment of zstor

This Python lib is responsible for deploying a zstor (0-stor) setup on packet.net devices.
Each service will be run on a separate packet device.

This lib is meant to build a benchmarking setup and is not meant for a production setup. It is also not recommended to run this on production packet.net projects.

## Prerequisites

* Have the development branch (master at time of writing still has a bug that prevents this script from working properly, specifically in core9) of Jumpscale (js9, core9, lib9 and prefab9) installed.
Installation guide can be found [here](https://github.com/Jumpscale/bash#to-install).
* Have a packet.net token
* Add an ssh key to your ssh-agent and have the public key on packet.net associated with your account and/or the project from your token.
* Don't have any hostnames with the following convention prior to running a deployment as they are used by this library:
    * zstordb[0-n]
    * zstormeta[0-n]
    * zstorbench0

## Deploy a zstordb

```python
packet_deploy = lib.SetupZstorPacket("insert your packet token here")

# Start new devices running zstordb.
packet_deploy.init_and_run_zstordb(
    servers=2,          # Amount of zstordb servers to be deployed
    facility='ams1',    # Packet.net facility to deploy on
    plan="baremetal_0", # Packet.net plan (device type)
    os="ubuntu_16_04",  # OS of the packet device
    no_auth=True,       # Defines if zstordb needs 
                        # to have authentication disabled
    jobs=0,             # Zstordb jobs flag (workers for grpc)
    port=1230,          # Zstordb port
    profile=None        # Type of profiling on zstordb (None is disabled)
    profile_dest="./profile",  # Destination folder where the profiling will be downloaded.
                        # inside that folder, it will create a folder with the hostname
                        # that has his profiling file inside of it.
    branch="master"     # Branch of 0-stor that will be installed
)

# Stop and removes the packet devices running zstordb.
# If profiling was enabled on the zstordb, it wil now be downloaded
# to the destination folder.
packet_deploy.stop_zstordb() 
```


## Deploy an etcd metadata server

```python
packet_deploy = lib.SetupZstorPacket("insert your packet token here")

# Start new devices running a metadata server (etcd).
packet_deploy.init_and_run_meta(
    servers=2,          # Amount of meta servers to be deployed
    facility='ams1',    # Packet.net facility to deploy on
    plan="baremetal_0", # Packet.net plan (device type)
    os="ubuntu_16_04",  # OS of the packet device
    etcd_version="3.2.13" # Version of etcd to be installed
    client_port=1200,   # port for clients to connect with etcd
    peer_port=1300,     # peer port for etcd (for etcd clusters)
)

# Remove all metadata server devices that have been set up.
stop_meta()
```

## Deploy and run zstorbench

```python
packet_deploy = lib.SetupZstorPacket("insert your packet token here")

# Start a new device and install zstorbench on it
packet_deploy.init_zstorbench(
    facility='ams1',    # Packet.net facility to deploy on
    plan="baremetal_0", # Packet.net plan (device type)
    os="ubuntu_16_04",  # OS of the packet device
    branch="master",    # Branch of 0-stor that will be installed
)

# Run zstorbench
# Result and profile will be downloaded after the benchmarking has finished
packet_deploy.run_zstorbench(
    config="./config.yaml", # Location of the config file for zstorbench
    out="./result.yaml",    # Destination file of the zstorbench result
    profile=None,           # Type of profiling on zstorbench (None is disabled)
    profile_dest="./profile",  # Destination folder where the profiling will be downloaded.
)

# Remove the device running zstorbench
stop_zstorbench()
```
