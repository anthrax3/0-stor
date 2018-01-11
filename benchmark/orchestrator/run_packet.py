from lib import SetupZstorPacket
from lib import Config
from threading import Thread
import sys
import time

def main(argv):
    facility = 'nrt1'
    #facility = 'sjc1'
    #facility = 'ams1'
    #facility = 'ewr1'
    token = "qokxuHhWpKURQsrPtib76BUF9p8wt67R"

    #zstordb(token=token, facility=facility)
    #zstorbench(token=token, facility=facility)
    #meta(token=token, facility=facility)
    #run_full(token=token, facility=facility)
    run_full_async(token=token, facility=facility)

def run_full_async(token="", facility=""):
    packet = SetupZstorPacket(token)

    t_zstordb = Thread(target=packet.init_and_run_zstordb, kwargs={'servers':4, 'facility':facility, 'branch':"packets_beta"})
    t_zstordb.start()

    t_zstormeta = Thread(target=packet.init_and_run_meta, kwargs={'servers':2, 'facility':facility})
    t_zstormeta.start()

    t_zstorbench = Thread(target=packet.init_zstorbench, kwargs={'branch':"benchmark_orchestrator_beta2", 'facility':facility})
    t_zstorbench.start()

    t_zstordb.join()
    t_zstormeta.join()
    t_zstorbench.join()

    # run zstorbench
    while True:
        print("data shards:", packet.zstordb_addresses)
        print("meta shards: ", packet.meta_addresses)
        resp = input("\nType 'exit' to stop benchmarking, press enter to run a benchmark:\n")
        if resp == "exit":
            break
        packet.run_zstorbench(config="./scenariosConf.yaml", out="./results/result.yaml")

    # stop all the things
    packet.stop()

def run_full(token="", facility=""):
    packet = SetupZstorPacket(token)

    # run zstordbs (no profiling for now)
    db_ips = packet.init_and_run_zstordb(servers=2, facility=facility, branch="benchmark_orchestrator_beta2")

    # run meta servers
    meta_ips = packet.init_and_run_meta(servers=1, facility=facility)

    # put addresses into zstorbench config
    print("data shards:", db_ips)
    print("meta shards: ", meta_ips)
    input("Press enter to continue benchmarking")

    # run zstorbench
    packet.init_zstorbench(branch="benchmark_orchestrator_beta2", facility=facility)
    packet.run_zstorbench(config="./scenariosConf.yaml", out="./results/result.yaml")

    # stop all the things
    packet.stop()

def zstordb(token="", facility=""):
    print("init")
    packet = SetupZstorPacket(token)

    print("deploying")
    ips = packet.init_and_run_zstordb(servers=2, profile="cpu", facility=facility, branch="master")
    print("deployed:", ips)

    #time.sleep(10)
    #print("shutting down")
    #packet.stop_zstordb()

    print("done")

def meta(token="", facility=""):
    packet = SetupZstorPacket(token)

    ips = packet.init_and_run_meta(servers=1, facility=facility)
    print(ips)

    time.sleep(10)

    packet.stop_meta()
    
def zstorbench(token="", facility=""):
    print("init")
    packet = SetupZstorPacket(token)

    print("deploying")
    packet.init_zstorbench(branch="benchmark_orchestrator_beta3", facility=facility)

    packet.run_zstorbench(config="./scenariosConf.yaml")
    print("done")

def stopall(token=""):
    packet = SetupZstorPacket()
    packet.stop()

if __name__ == '__main__':
    main(sys.argv[1:])