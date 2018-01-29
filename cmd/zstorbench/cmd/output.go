package cmd

import (
	"io/ioutil"

	"github.com/zero-os/0-stor/benchmark/bencher"
	"github.com/zero-os/0-stor/benchmark/config"

	yaml "gopkg.in/yaml.v2"
)

// NewOutputFormat returns a new OutputFormat
func NewOutputFormat() OutputFormat {
	var o OutputFormat
	o.Scenarios = make(map[string]ScenarioOutputFormat)
	return o
}

// OutputFormat represents the output format of a full benchmark
type OutputFormat struct {
	Scenarios map[string]ScenarioOutputFormat
}

//ScenarioOutputFormat represents a scenario result for outputting
type ScenarioOutputFormat struct {
	Results      []*bencher.Result `yaml:"results,omitempty"`
	ScenarioConf config.Scenario   `yaml:"scenario,omitempty"`
	Error        string            `yaml:"error,omitempty"`
}

//FormatOutput formats the output of the benchmarking scenario
func FormatOutput(results []*bencher.Result, scenarioConfig config.Scenario, err error) ScenarioOutputFormat {
	var output ScenarioOutputFormat
	output.ScenarioConf = scenarioConfig
	if err != nil {
		output.Error = err.Error()
		return output
	}
	output.Results = results
	return output
}

// writeOutput writes OutputFormat to provided file
func writeOutput(filePath string, output OutputFormat) error {
	yamlBytes, err := yaml.Marshal(output)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, yamlBytes, 0644)
}
