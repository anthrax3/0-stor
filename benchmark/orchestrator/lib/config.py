# Copyright (C) 2017-2018 GIG Technology NV and Contributors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
    Package config includes functions to set up configuration for benchmarking scenarios
"""
import time
import os
from re import split
from copy import deepcopy
from subprocess import check_output
from threading import Thread
import yaml
from lib.zstor_local_setup import SetupZstor
from lib.zstor_packet_setup import SetupZstorPacket

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
PARAMETERS_DICT = {'encryption': {'type', 'private_key'},
                   'compression': {'type', 'mode'}}

PROFILES = {'cpu', 'mem', 'trace', 'block'}

LOCAL_DEPLOYMENT = 'local'
PACKETS_DEPLOYMENT = 'packet.net'
DEFAULT_BRANCH = 'master'

DEFAULT_PACKETS_CONFIG = {  'facility': 'ams1',
                            'plan': 'baremetal_0',
                            'etcd_version': '3.2.13',
                            'os': 'ubuntu_16_04'}
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

        # extract branch
        self.branch = config.get('branch', DEFAULT_BRANCH)

        # check packet.net config, if not given set up local deployment
        self.packets = config.get('packet.net', None)
        if self.packets:
            self.deployment = PACKETS_DEPLOYMENT
            self.deploy = SetupZstorPacket()

            # set default options for packet.net config
            for key, val in DEFAULT_PACKETS_CONFIG.items():
                parameter = self.packets.get(key, None)
                if not parameter or parameter == 'default':
                    self.packets.update({key: val})

        else:
            self.deployment = LOCAL_DEPLOYMENT
            self.deploy = SetupZstor()

    def new_profile_dir(self, path=""):
        """
        Create new directory for profile information in given path and dump current config
        Returns zstordb profile dir and client profile dir
        """

        if self.profile:
            directory = os.path.join(path, 'profile_information')
            if not os.path.exists(directory):
                os.makedirs(directory)
            directory = os.path.join(directory, 'profile_' + str(self.count_profile))
            if not os.path.exists(directory):
                os.makedirs(directory)

            file = os.path.join(directory, 'config.yaml')
            with open(file, 'w+') as outfile:
                yaml.dump({'scenarios': {'scenario': self.template}}, 
                            outfile, 
                            default_flow_style=False, 
                            default_style='')

            zstordb_dir = os.path.join(directory, 'zstordb')
            zstor_client_dir = os.path.join(directory, 'zstorclient')

            self.count_profile += 1
            return zstordb_dir, zstor_client_dir
        return "", "" 

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

    def deploy_zstor(self, profile_dir="profile_zstordb"):
        """
        Run zstordb and etcd servers
        Profile parameters will be used for zstordb
        """

        self.update_deployment_config()

        if self.deployment == LOCAL_DEPLOYMENT:
            self.deploy.run_data_shards(servers=self.data_shards_nr,
                                            no_auth=self.no_auth,
                                            jobs=self.zstordb_jobs)
            self.deploy.run_meta_shards(servers=self.meta_shards_nr)            
            self.wait_local_servers_to_start()            

        if self.deployment == PACKETS_DEPLOYMENT:
            t_data_shards = Thread(target=self.deploy.run_data_shards,
                                   kwargs={'servers' : self.data_shards_nr,
                                            'no_auth' : self.no_auth,
                                            'jobs' : self.zstordb_jobs,
                                            'facility' : self.packets['facility'],
                                            'plan' : self.packets['plan'],
                                            'os' : self.packets['os'],
                                            'profile' : self.profile,
                                            'profile_dest' : profile_dir})
            t_data_shards.start()            

            t_meta_shards = Thread(target=self.deploy.run_meta_shards,
                                   kwargs={'servers' : self.meta_shards_nr,
                                            'facility' : self.packets['facility'],
                                            'plan' :self.packets['plan'],
                                            'os' : self.packets['os'],
                                            'etcd_version' : self.packets['etcd_version']})
            t_meta_shards.start()

            t_benchmark = Thread(target=self.deploy.init_benchmark,
                                    kwargs={'branch' : self.branch, 
                                            'facility' : self.packets['facility'],
                                            'plan' :self.packets['plan'],
                                            'os' : self.packets['os']})
            t_benchmark.start()
            
            t_data_shards.join()
            t_meta_shards.join()
            t_benchmark.join()
                                                
        self.datastor.update({'shards': self.deploy.data_shards})
        self.metastor.update({'shards': self.deploy.meta_shards})


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
    def run_benchmark(self, config='config.yaml', out='result.yaml', profile_dir='./profile'):
        """ Runs benchmarking """

        self.deploy.run_benchmark(config=config,
                                    out=out,
                                    profile=self.profile,
                                    profile_dir=profile_dir)

    def stop_zstor(self):
        """ Stop zstordb and datastor servers """
        self.deploy.stop()

        self.datastor.update({'shards': []})
        self.metastor.update({'db':{'endpoints': []}})       

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