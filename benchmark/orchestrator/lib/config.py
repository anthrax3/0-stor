"""
    Package config includes functions to set up configuration for benchmarking scenarios
"""
import time
import os
from re import split
from copy import deepcopy
from subprocess import check_output
import yaml
from lib.zstor_local_setup import SetupZstor

class InvalidBenchmarkConfig(Exception):
    pass

# list of supported benchmark parameters
PARAMETERS = {'block_size',
              'key_size',
              'value_size',
              'clients',
              'method',
              'block_size',
              'data_shards',
              'parity_shards',
              'meta_shards_nr',
              'zstordb_jobs'}
PARAMETERS_DICT = {'encryption': 'type',
                   'compression': {'type', 'mode'}}

PROFILES = {'cpu', 'mem', 'trace', 'block'}

class Config:
    """
    Class Config includes functions to set up environment for the benchmarking:
        - deploy zstor servers
        - config zstor benchmark client
        - iterate over range of benchmark parameters

    @template contains zerostor client config
    @benchmark defines iterator over provided benchmarks
    """
    def __init__(self, config_file):
        # read config yaml file
        with open(config_file, 'r') as stream:
            try:
                config = yaml.load(stream)
            except yaml.YAMLError as exc:
                raise exc
        # fetch template config for benchmarking
        self._template0 = config.get('template', None)
        self.restore_template()

        # fetch bench_config from template
        bench_config = self.template.get('bench_config', None)
        if not bench_config:
            raise InvalidBenchmarkConfig('no bench_config given in template')
        self.zstordb_jobs = bench_config.get('zstordb_jobs', 0)

        if not self.template:
            raise InvalidBenchmarkConfig('no zstor config given')

        # extract benchmarking parameters
        self.benchmark = iter(self.benchmark_generator(config.pop('benchmarks', None)))
        # extract profiling parameter
        self.profile = config.get('profile', None)

        if self.profile and (self.profile not in PROFILES):
            raise InvalidBenchmarkConfig("profile mode '%s' is not supported"%self.profile)

        self.count_profile = 0

        self.deploy = SetupZstor()

    def new_profile_dir(self, path=""):
        """
        Create new directory for profile information in given path and dumps current config
        """
        if self.profile:
            directory = '%s/profile_information'%path
            if not os.path.exists(directory):
                os.makedirs(directory)
            directory = '%s/profile_%s'%(directory,str(self.count_profile))         
            if not os.path.exists(directory):
                os.makedirs(directory)
            file = "%s/config.yaml"%directory
            with open(file, 'w+') as outfile:
                yaml.dump({'scenarios': {'scenario': self.template}}, 
                            outfile, 
                            default_flow_style=False, 
                            default_style='')             
            self.count_profile += 1    
            return directory
        return "" 

    def benchmark_generator(self,benchmarks):
        """
        Iterate over list of benchmarks
        """     
        if benchmarks:
            for bench in benchmarks:
                yield BenchmarkPair(bench)        
        else:
            yield BenchmarkPair()

    def alter_template(self, key_id, val): 
        """
        Recurcively search and ppdate @id config field with new value @val
        """
        def replace(d, key_id, val):
            for key in list(d.keys()):
                v = d[key]
                if isinstance(v, dict):
                    if isinstance(key_id, dict):
                        if key == list(key_id.items())[0][0]:
                            return replace(v, key_id[key], val)
                    if replace(v, key_id, val):
                        return True
                else:
                    if key == key_id:
                        parameter_type = type(d[key])
                        try:
                            d[key] = parameter_type(val)
                        except:
                            raise InvalidBenchmarkConfig("for '{}' cannot convert val = {} to type {}".format(key,val,parameter_type))
                        return True
            return False
        if not replace(self.template, key_id, val):
            raise InvalidBenchmarkConfig("parameter %s is not supported"%key_id)

    def restore_template(self):
        """ Restore initial zstor config """

        self.template = deepcopy(self._template0)

    def save(self, file_name):
        """ Save current config to file """

        # prepare config for output
        output = {'scenarios': {'scenario': self.template}}

        # write scenarios to a yaml file
        with open(file_name, 'w+') as outfile:
            yaml.dump(output, outfile, default_flow_style=False, default_style='')

    def update_deployment_config(self):
        """ 
        Fetch current zstor server deployment config
                ***specific for beta2***
        """
        try:
            self.datastor =  self.template['zstor_config']['datastor']
            distribution = self.datastor['pipeline']['distribution']
            self.data_shards_nr=distribution['data_shards'] + distribution['parity_shards']
        except:
            raise InvalidBenchmarkConfig("distribution config is not correct")
        try:
            self.metastor  = self.template['zstor_config']['metastor']
            self.meta_shards_nr = self.metastor['meta_shards_nr']
        except:
            raise InvalidBenchmarkConfig("number of metastor servers is not given")
        
        self.no_auth = True
        IYOtoken = self.template['zstor_config'].get('iyo', None)
        if IYOtoken:
            self.no_auth = False

    def deploy_zstor(self):
        """ Run zstordb and etcd servers """

        self.deploy.run_zstordb_servers(servers=self.data_shards_nr,
                                        no_auth=self.no_auth,
                                        jobs=self.zstordb_jobs)
        self.deploy.run_etcd_servers(servers=self.meta_shards_nr)

        self.datastor.update({'shards': self.deploy.data_shards})
        self.metastor.update({'db':{'endpoints': self.deploy.meta_shards}})

    def stop_zstor(self):
        """ Stop zstordb and etcd servers """

        self.deploy.stop_etcd_servers()
        self.deploy.stop_zstordb_servers()
        self.deploy.cleanup()

        self.datastor.update({'shards': []})
        self.metastor.update({'db':{'endpoints': []}})



    def wait_local_servers_to_start(self):
        """ Check whether ztror and etcd servers are listening on the ports """

        addrs = self.deploy.data_shards + self.deploy.meta_shards
        servers = 0
        timeout = time.time() + 20
        while servers < len(addrs):
            servers = 0
            for addr in addrs:
                port = ':%s'%split(':', addr)[-1]
                try:
                    responce = check_output(['lsof', '-i', port])
                except:
                    responce=0
                if responce:
                    servers += 1
                if time.time() > timeout:
                    raise TimeoutError("couldn't run all required servers. Check that ports are free")

class Benchmark():
    """ Benchmark class is used defines and validates benchmark parameter """

    def __init__(self, parameter={}):
       
        if parameter:
            self.id = parameter.get('id', None)
            self.range = parameter.get('range', None)
            if not self.id or not self.range:
                raise InvalidBenchmarkConfig("parameter id or range is missing")
            
            if isinstance(self.id, dict):
                def contain(d, id):
                    if isinstance(d, dict) and isinstance(id, dict):
                        for key in list(d.keys()):
                            if id.get(key, None):
                                if contain(d[key], id[key]):
                                    return True
                    else:    
                        if id in d:
                            return True
                    return False
                if not contain(PARAMETERS_DICT, self.id):
                    raise InvalidBenchmarkConfig("parameter {0} is not supported".format(self.id))
            else:
                if self.id not in PARAMETERS:
                    raise InvalidBenchmarkConfig("parameter {0} is not supported".format(self.id))
             
            
            try:
                self.range = split("\W+", self.range)
            except:
                pass

        else:
            # return empty Benchmark
            self.range = [' ']
            self.id = ''

    def empty(self):
        """ Return True if benchmark is empty """
        if (len(self.range) == 1) and not self.id:
            return True
        return False
class BenchmarkPair():
    """
    BenchmarkPair defines primary and secondary parameter for benchmarking
    """
    def __init__(self, bench_pair={}):
        if bench_pair:
            # extract parameters from a dictionary
            self.prime = Benchmark(bench_pair.pop('prime_parameter', None))
            self.second = Benchmark(bench_pair.pop('second_parameter', None))

            if not self.prime.empty() and self.prime.id == self.second.id:
                raise InvalidBenchmarkConfig("primary and secondary parameters should be different")
            
            if self.prime.empty() and not self.second.empty():
                raise InvalidBenchmarkConfig("if secondary parameter is given, primary parameter has to be given")
        else:
            # define empty benchmark
            self.prime = Benchmark()
            self.second = Benchmark()                                                