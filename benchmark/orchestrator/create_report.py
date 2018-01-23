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

from lib import Output
import sys
from getopt import getopt

def main(argv):
    # default path to template yaml file
    input_data = "benchmark.yaml"

    # default path where config for scenarios is written
    output_dir = "report"

    # default path where orchestrator config in given
    output_dir = "templateConf.yaml"
    
    try:
        opts, args = getopt(argv,"i:o:c:",["","","",])

        # check if output directories are given
        for opt, arg in opts:
            if opt == '-i':
                # set new file for input
                input_data = arg
            if opt == '-o':
                # set new file for output
                output_dir = arg   
            if opt == '-c':
                # set new file for output
                conf_file = arg
                print("conf = ", conf_file)   
                       
    except:
        print("default paths are used")

    # parse output of the benchmarking
    output = Output(input_data, output_dir)

    # plot and save figures vs time
    output.plot_per_interval()

    # plot and save figures vs parameter IN PROGRESS
    output.plot_range()    

    output.create_md(output_dir, conf_file)
    

if __name__ == '__main__':
    main(sys.argv[1:])    