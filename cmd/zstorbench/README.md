# zstor benchmark client

Benchmark client provides tools for benchmarking and profiling zstor client for various scenarios.

Configuration for benchmarking scenarios should be given in YAML format (see [example](#benchmark-config) of the config file bellow).
Benchmarking program outputs results for all provided scenarios to a single output file in YAML format (see [example](#output file) of the output file bellow). 


## Getting started
Install all necessary `zstor` components by running
```bash
make install
```
In order to start a benchmarking run, all necessary [zstor servers](https://github.com/zero-os/0-stor/blob/master/docs/gettingstarted.md) have to be set up.

Start the benchmarking
```bash
zstorbench -C config.yaml --out-benchmark benchmark.yaml
```

`zstorbench` has the following options:
```
  -C, --conf string                path to a config file (default "zstorbench_config.yaml")
  -D, --debug                      Enable debug logging.
  -h, --help                       help for zstorbench
      --out-benchmark string       path and filename where benchmarking results are written (default "benchmark.yaml")
      --out-profile string         path where profiling files are written (default "./profile")
      --profile-mode profileMode   Enable profiling mode, one of [cpu, mem, block, trace]
```


Start benchmarking with optional input/output files
```bash
zstorbench --conf "input_config.yaml" --out-benchmark "output_benchmark.yaml"
```

Start benchmarking and profiling
```bash
zstorbench --out-profile "outputProfileInfo" --profile-mode cpu
```

## Benchmark config

Client config contains a list of scenarios. 
Each scenario is associated with a corresponding scenarioID and provides two sets of parameters: 
`zstor_config` and `bench_conf`.
Structure `zstor_config` are nessesary to create a `zstor client` and can be parsed into a type `client.Policy` of [zstor client package](https://github.com/zero-os/0-stor/tree/master/client). 

`iyo` has to be given in case if `zstor` server is running with flag `--no-auth` and, therefore, require no authentification via `itsyou.online`.
Note, that invalid `ioy` token lead to an error even if authentification is unset.

Structure `bench_conf` represents such benchmark parameters like duration of the performance test, maximum number of operations, maximum benchmark duration and output format.
One of two parameters `duration` and `operations` has to be provided. If both are given, the benchmarking program terminates as soon as one of the following events occurs:
 + number of executed operations reached `operations`
 + timeout set by `duration` elapsed

`method` defines which operation is benchmarked:
 + `read` - for reading from zstor
 + `write` - for writing to zstor

`result_output` specifies interval of the data collection (`perinterval` in the results) and can take values:  
 + per_second
 + per_minute
 + per_hour

if empty or invalid, there will be no interval data collection.

The following example of a config file represents a benchmarking scenario `bench1`.

``` yaml
scenarios:
  bench1:
    zstor_config:
      iyo:  # if empty or omitted, the zstordb servers set up for the benchmark 
            # need to be run with the no-auth flag.
            # For benching with authentication, provide it with valid itsyou.online credential
        organization: "bench_org"
        app_id: "an_iyo_bench_app_id"
        app_secret: "an_iyo_bench_app_secret"
      namespace: adisk
      datastor:
        shards:
        - 127.0.0.1:45627
        - 127.0.0.1:49861
        - 127.0.0.1:37355
      metastor:
          shards: # If empty or omitted, an in memory metadata server will be used
                  # Otherwise it will presume to have etcd servers running on these addresses
            - 127.0.0.1:1300
            - 127.0.0.1:1301
      pipeline:
        block_size: 4096
        compression:
          type: snappy # snappy is the default, other options: lz4, gzip
          mode: default # default is the default, for gzip other options: best_speed, best_compression
        encryption:
          private_key: ab345678901234567890123456789012
        distribution:
          data_shards: 2
          parity_shards: 1
    bench_config:
      method: write
      result_output: per_second # if empty or invalid, perinterval in the result will be empty
      duration: 5
      operations: 0
      key_size: 48
      value_size: 128
      clients: 1 # number of concurrent clients
```

## Output file

Benchmarking program writes results of the performance tests to an output file. All scenario configuration is collected in `scenario`. All numerical results can be fetched from `results`.

``` yaml
scenarios:
  bench1:
    results:
    - count: 367            # number of reads/writes performed during the benchmark
      duration: 5.0093327  # total duration of the benchmark
      perinterval:         # number of reads/writes per time unit
      - 84
      - 64
      - 67
      - 70
      - 82
    scenario:         # zstor config used for the benchmark
      zstor_config:
      namespace: mynamespace # itsyou.online namespace
      iyo:  # itsyou.online authentification token
        organization: myorg  # itsyou.online organization name
        app_id: appID        # itsyou.online Application ID
        app_secret: secret   # itsyou.online Secret
        datastor:
          shards:
          - 127.0.0.1:45627
          - 127.0.0.1:49861
          - 127.0.0.1:37355
        metastor:
          shards:
          - 127.0.0.1:1300
          - 127.0.0.1:1301
        pipeline:
          block_size: 4096
          hashing:
            type: blake2b_256
            private_key: ""
          compression:
            mode: default
            type: snappy
          encryption:
            private_key: ab345678901234567890123456789012
            type: aes
          distribution:
            data_shards: 2
            parity_shards: 1
      bench_config:
        method: write
        result_output: per_second
        duration: 5
        operations: 0
        clients: 1
        key_size: 48
        value_size: 128
```