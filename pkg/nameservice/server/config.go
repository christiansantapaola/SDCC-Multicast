package server

import (
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"time"
)

/*
	Configurazione del name server.
	DialTimeout: time out per la connesione al cluster etcd
	Endpoints: punti di accesso al cluster etcd.
*/

type Config struct {
	Etcd struct {
		DialTimeout time.Duration `yaml:"dial_timeout"`
		Endpoints   []string      `yaml:"endpoints"`
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
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	etcd := make([]string, 0)
	etcdOsVar := os.Getenv("SDCC_NAME_SERVER_ETCD")
	if etcdOsVar == "" {
		etcd = append(etcd, "127.0.0.1:2379")
		etcd = append(etcd, "127.0.0.1:2380")
	} else {
		addrs := strings.Split(etcdOsVar, ";")
		for _, addr := range addrs {
			etcd = append(etcd, addr)
		}
	}
	cfg := Config{}
	cfg.Etcd.DialTimeout = 5 * time.Second
	cfg.Etcd.Endpoints = etcd
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
