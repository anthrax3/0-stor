package config

import (
	"fmt"

	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v2"

	"github.com/zero-os/0-stor/client/lib/block"
	"github.com/zero-os/0-stor/client/lib/chunker"
	"github.com/zero-os/0-stor/client/lib/compress"
	"github.com/zero-os/0-stor/client/lib/distribution"
	"github.com/zero-os/0-stor/client/lib/encrypt"
	"github.com/zero-os/0-stor/client/lib/hash"
	"github.com/zero-os/0-stor/client/lib/replication"
)

// Pipe defines each 0-stor client pipe
type Pipe struct {
	Name string `yaml:"name"`

	// Type of this pipe, must be one of:
	// chunker, compress, distribution, encrypt, hash, replication
	Type string `yaml:"type" validate:"nonzero"`

	Config interface{} `yaml:"config"`
}

// CreateBlockReader creates a block reader
func (p Pipe) CreateBlockReader(shards []string, org, namespace string) (block.Reader, error) {
	switch p.Type {
	case compressStr:
		conf := p.Config.(compress.Config)
		return compress.NewReader(conf)
	case encryptStr:
		conf := p.Config.(encrypt.Config)
		return encrypt.NewReader(conf)
	case distributionStr:
		conf := p.Config.(distribution.Config)
		return distribution.NewStorRestorer(conf, shards, org, namespace)
	default:
		return nil, fmt.Errorf("invalid type:%v", p.Type)
	}
}

// CreateBlockWriter creates block writer
func (p Pipe) CreateBlockWriter(w block.Writer, shards []string, org, namespace string) (block.Writer, error) {
	switch p.Type {
	case compressStr:
		conf := p.Config.(compress.Config)
		return compress.NewWriter(conf, w)
	case distributionStr:
		conf := p.Config.(distribution.Config)
		return distribution.NewStorDistributor(conf, shards, org, namespace)
	case encryptStr:
		conf := p.Config.(encrypt.Config)
		return encrypt.NewWriter(w, conf)
	default:
		return nil, fmt.Errorf("unsupported type:%v", p.Type)
	}
}

// set p.Config to proper type.
// because by default the parser is going to
// set tyhe type as map[string]interface{}
func (p *Pipe) setConfigType() error {
	b, err := yaml.Marshal(p.Config)
	if err != nil {
		return err
	}

	switch p.Type {
	case chunkerStr:
		var conf chunker.Config
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return err
		}
		p.Config = conf

	case compressStr:
		var conf compress.Config
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return err
		}
		p.Config = conf

	case encryptStr:
		var conf encrypt.Config
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return err
		}
		p.Config = conf

	case hashStr:
		var conf hash.Config
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return err
		}
		p.Config = conf

	case replicationStr:
		var conf replication.Config
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return err
		}
		p.Config = conf

	case distributionStr:
		var conf distribution.Config
		if err := yaml.Unmarshal(b, &conf); err != nil {
			return err
		}
		p.Config = conf
	default:
		panic("invalid type:" + p.Type)

	}
	return nil
}

// validate a pipe config
func (p Pipe) validate() error {
	if err := validator.Validate(p); err != nil {
		return err
	}

	if _, ok := validPipes[p.Type]; !ok {
		return fmt.Errorf("invalid pipe: %v", p.Type)
	}

	return nil
}
