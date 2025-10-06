package indexer

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type yamlConfig struct {
	Networks map[string]struct {
		RPC        string `yaml:"rpc"`
		Identity   string `yaml:"identity"`
		Reputation string `yaml:"reputation"`
		Validation string `yaml:"validation"`
	} `yaml:"networks"`
}

func loadConfig(path string) ([]Chain, error) {
	if path == "" {
		return []Chain{}, nil
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return []Chain{}, err
	}
	var cfg yamlConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return []Chain{}, err
	}
	out := make([]Chain, 0, len(cfg.Networks))
	for name, n := range cfg.Networks {
		c := Chain{
			Name:       name,
			RPC:        os.ExpandEnv(n.RPC),
			Identity:   n.Identity,
			Reputation: n.Reputation,
			Validation: n.Validation,
		}
		// 👇 add a log entry here
		log.WithFields(log.Fields{
			"chain":      c.Name,
			"rpc":        c.RPC,
			"identity":   c.Identity,
			"reputation": c.Reputation,
			"validation": c.Validation,
		}).Info("loaded network config")

		out = append(out, c)
	}
	return out, nil
}
