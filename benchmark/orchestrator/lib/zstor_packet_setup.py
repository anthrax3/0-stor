from js9 import j
import packet, sys, time
from lib.zstor_local_setup import is_profile_flag
from threading import Thread, Lock
from re import split

# temp dir on packet device
_TMP_DIR = '/tmp'
_PROF_DIR = _TMP_DIR + "/zstorprof"
_ZSTORBENCH_HOSTNAME = "zstorbench0"
_ZSTORBENCH_CONF = _TMP_DIR + "/zstorbenchconf/config.yaml"
_ZSTORBENCH_OUT = _TMP_DIR  + "/zstorbenchconf"
_DATA_DIR = _TMP_DIR + "/data"
_META_DIR = _TMP_DIR + "/meta"
_ETCD_DIR = _TMP_DIR + "/etcd"

class SetupZstorPacket:
    """
    SetupZstorPacket is responsible for managing a zstor setup on packet.net devices.
    The lib presumes that the os on the devices will be Linux (Ubuntu).

    SetupZstorPacket will use the following hostname convention for the packet.net devices:
        zstordb[0-n]
        zstormeta[0-n]
        zstorbench0
    """
    def __init__(self):

        self.p_client = j.clients.packetnet.get()

        self._meta_devices = {}
        self._meta_ips = {}
        self._meta_lock = Lock()
        self.meta_shards = []

        self._zstordb_devices = {}
        self._zstordb_lock = Lock()
        self.zstordb_prof = False
        self.zstordb_prof_dest = "."
        self.data_shards = []

        self.zstorbench_prefab = None

    def run_data_shards(self, servers=2, 
                        facility='ams1', 
                        plan="baremetal_0", 
                        os="ubuntu_16_04",
                        no_auth=True, 
                        jobs=0, 
                        port=1230, 
                        profile=None, 
                        profile_dest="./zstordb_profile",
                        branch="master"):
        """
        Runs zstordb's on packet.net devices.
        Returns list of ip addresses with ports where zstordb is running,
        can also be retrieved from SetupZstorPacket.zstordb_addresses
        """
        self.zstordb_prof_dest = profile_dest
        if profile != None:
            self.zstordb_prof = True
            self.profile_dest = profile_dest

        ts = []
        for i in range(servers):
            t = Thread(target=self._setup_new_zstordb_machine,\
            args=[i, plan, os, facility, port, no_auth, jobs, profile, branch])

            t.start()
            ts.append(t)

        # Wait for threads to complete
        for t in ts:
            t.join()

    def _setup_new_zstordb_machine(self, i, plan, os, facility, port, no_auth, jobs, profile, branch):
        """
        Creates new packet device and installs zstordb onto it.
        Is designed to run in a separate thread
        """
        name = "zstordb" + str(i)
        node = self.p_client.startDevice(hostname=name, plan=plan, os=os, facility=facility)
        ip = node.addr

        self._zstordb_lock.acquire()
        self.data_shards.append(ip + ":" + str(port))
        self._zstordb_lock.release()

        install_zstor(node.prefab, branch)

        # run zstordb
        node.prefab.core.dir_ensure(_DATA_DIR)
        node.prefab.core.dir_ensure(_META_DIR)
        cmd = "zstordb --listen %s --data-dir %s --meta-dir %s --jobs %s" \
              % (":" + str(port), _DATA_DIR, _META_DIR, str(jobs))
        if no_auth:
            cmd += " --no-auth"
        if profile and is_profile_flag(profile):
            profdir = _PROF_DIR + "/" + name
            node.prefab.core.dir_ensure(profdir)
            cmd += " --profile-mode %s" % profile
            cmd += " --profile-output %s" % profdir

        # run zstordb in background
        cmd += " &>/dev/null &"
        node.prefab.core.execute_bash(cmd)

        self._zstordb_lock.acquire()
        self._zstordb_devices[name] = node.prefab
        self._zstordb_lock.release()

    def stop_data_shards(self,):
        """
        Stop all packet.net devices running a zstordb.
        Retrieve profile result if required.
        """
        for hostname, prefab in self._zstordb_devices.items():

            # try and fetch profiling if needed
            if self.zstordb_prof:
                # terminate zstordb
                prefab.core.execute_bash("pkill -SIGINT zstordb", die=False)

                # download files to destination dir
                profdir = _PROF_DIR + "/" + hostname
                j.tools.prefab.local.core.dir_ensure(self.zstordb_prof_dest)
                prefab.core.download(profdir, self.zstordb_prof_dest)

            self.p_client.removeDevice(hostname)

    def run_meta_shards(self, servers=1, 
                              facility='ams1', 
                              plan="baremetal_0", 
                              os="ubuntu_16_04",
                              etcd_version="3.2.13", 
                              client_port=1200, 
                              peer_port=1300):
        """
        Run etcd metadata server(s) on new packet.net devices
        Returns list of ip addresses with ports where etcd is running,
        can also be retrieved from SetupZstorPacket.meta_addresses
        """
        init_cluster = ""
        ts = []

        # setup and install etcd on packet machines concurrently
        for i in range(servers):
            t = Thread(target=self._setup_new_meta_machine, args=[i, plan, os, facility, etcd_version])
            t.start()
            ts.append(t)

        # Wait for threads to complete
        for t in ts:
            t.join()

        # build data to run etcd
        for name, ip in self._meta_ips.items():
            self.meta_shards.append(ip + ":" + str(client_port))
            init_cluster += name + "=http://"+ ip + ":" + str(peer_port) + ","

        for i, (hostname, prefab) in enumerate(self._meta_devices.items()):
            # run etcd
            run_etcd(prefab, hostname, self._meta_ips[hostname], peer_port, client_port, init_cluster)

    def _setup_new_meta_machine(self,i, plan, os, facility, etcd_version):
        """
        Creates new packet device and installs and runs etcd onto it.
        Is designed to run in a separate thread
        """
        name = "zstormeta" + str(i)
        node = self.p_client.startDevice(hostname=name, plan=plan, os=os, facility=facility)
        ip = node.addr

        self._meta_lock.acquire()
        self._meta_ips[name] = ip
        self._meta_lock.release()
        
        install_etcd(node.prefab, version=etcd_version)
        self._meta_lock.acquire()
        self._meta_devices[name] = node.prefab
        self._meta_lock.release()

    def stop_meta_shards(self,):
        """Stop all packet.net device running etcd"""
        for hostname, prefab in self._meta_devices.items():
            pass
            self.p_client.removeDevice(hostname)

    def init_benchmark(self, 
                        branch="master", 
                        facility='ams1', 
                        plan="baremetal_0",
                        os="ubuntu_16_04"):
        """
        Sets up a zstorbench on a packet device
        """
        
        node = self.p_client.startDevice(\
            hostname=_ZSTORBENCH_HOSTNAME, plan=plan, os=os, facility=facility)

        # install zstorbench
        install_zstor(node.prefab, branch)

        self.zstorbench_prefab = node.prefab

    def run_benchmark(self, config="./config.yaml", out="result.yaml",\
    profile=None, profile_dest="./zstor_client_profile/"):
        """
        Start a zstorbench benchmark.
        Make sure to run start_zstorbench before calling this.
        """
        prefab = self.zstorbench_prefab
        
        # load bench config
        prefab.core.upload(config, _ZSTORBENCH_CONF)

        # run benchmark
        out_name = split('/', out)[-1]
        out_path_local = split('/', out)[:-1]
        if not out_path_local:
            out_path_local = '.'
        full_path_out = '%s/%s'%(_ZSTORBENCH_OUT, out_name)

        cmd = "zstorbench --conf %s --out-benchmark %s" % (_ZSTORBENCH_CONF, full_path_out)
        if profile != None:
            cmd += " --profile-mode %s --out-profile %s" % (profile, _PROF_DIR)

        prefab.core.execute_bash(cmd)

        # download results and profiling if required
        j.tools.prefab.local.core.dir_ensure(_ZSTORBENCH_OUT)
        if not j.tools.prefab.local.core.file_exists(full_path_out):
            raise RuntimeError('%s is not found'%full_path_out)

        prefab.core.download(full_path_out, out_path_local)
        if profile != None:
            j.tools.prefab.local.core.dir_ensure(profile_dest)
            prefab.core.download(_PROF_DIR, profile_dest)

    def stop_benchmark(self):
        """
        stops/deletes the machine zstorbench is running on
        """
        self.p_client.removeDevice(_ZSTORBENCH_HOSTNAME)
    def stop(self):
        self.stop_data_shards()
        self.stop_meta_shards()
        self.stop_benchmark()

def install_zstor(prefab, branch="master"):
    """
    Installs 0-stor on prefab device
    """
    prefab.system.package.install("git")
    prefab.runtimes.golang.install()
    # kill zstordb in case there is one running 
    prefab.core.execute_bash("pkill -SIGINT zstordb", die=False)
    cmd = """
    mkdir -p $GOPATH/src/github.com/zero-os/
    cd $GOPATH/src/github.com/zero-os
    rm -rf 0-stor
    git clone -b %s https://github.com/zero-os/0-stor.git
    cd 0-stor
    make install
    """ % (branch)
    prefab.core.execute_bash(cmd)

def install_etcd(prefab, version="3.2.13"):
    """
    Installs etcd on prefab device
    """
    source = "https://github.com/coreos/etcd/releases/download"
    
    cmd = """
    rm -f /tmp/etcd-v%s-linux-amd64.tar.gz
    rm -rf %s && mkdir -p %s
    curl -L %s/v%s/etcd-v%s-linux-amd64.tar.gz -o /tmp/etcd-v%s-linux-amd64.tar.gz
    tar xzvf /tmp/etcd-v%s-linux-amd64.tar.gz -C %s --strip-components=1
    rm -f /tmp/etcd-v%s-linux-amd64.tar.gz
    """ % (version,\
    _ETCD_DIR, _ETCD_DIR,\
    source, version, version, version,\
    version, _ETCD_DIR,\
    version)

    prefab.core.execute_bash(cmd)

def run_etcd(prefab, name, public_ip, peer_port, client_port, initial_cluster):
    """
    Runs etcd with cluster params provided
    """
    exe = _ETCD_DIR + "/etcd"
    public_ip = "http://" + public_ip
    local_ip = "http://127.0.0.1"

    local_client_addr = local_ip + ":" + str(client_port)
    public_client_addr = public_ip + ":" + str(client_port)
    public_peer_addr = public_ip + ":" + str(peer_port)

    cmd = """
    %s \\
    --name %s \\
    --initial-advertise-peer-urls %s \\
    --listen-peer-urls %s \\
    --listen-client-urls %s \\
    --advertise-client-urls %s \\
    --initial-cluster-state new \\
    --initial-cluster-token etcd-bench-cluster-1 \\
    --initial-cluster %s \\
    """ % (exe, name, public_peer_addr, public_peer_addr,\
    public_client_addr + "," + local_client_addr, public_client_addr, initial_cluster)
    # run meta db in background
    cmd += "&>/dev/null &"

    print(cmd)

    prefab.core.execute_bash(cmd)
