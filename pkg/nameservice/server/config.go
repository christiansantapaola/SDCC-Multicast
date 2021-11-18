package server

import (
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Config struct {
	Etcd struct {
		DialTimeout time.Duration 	`yaml:"dial_timeout"`
		Endpoints []string			`yaml:"endpoints"`
	}
}

func ReadCfg(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func GenDefaultCfg(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg := Config{}
	cfg.Etcd.DialTimeout = 5 * time.Second
	cfg.Etcd.Endpoints = []string{"localhost:2379", "localhost:2380"}
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
