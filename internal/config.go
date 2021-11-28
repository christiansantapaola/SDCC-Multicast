package internal

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	NameServerAddress string        `yaml:"name_server_address"`
	UserId            string        `yaml:"user_id"`
	UserIp            string        `yaml:"user_ip"`
	UserPort          int           `yaml:"user_port"`
	Secure            bool          `yaml:"secure"`
	Verbose           bool          `yaml:"verbose"`
	Timeout           time.Duration `yaml:"timeout"`
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
	cfg := Config{
		NameServerAddress: "127.0.0.1:2080",
		UserId:            "",
		UserIp:            "127.0.0.1",
		UserPort:          2079,
		Secure:            false,
		Verbose:           false,
		Timeout:           time.Duration(5 * time.Second),
	}
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}

func Update(path string, cfg *Config) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
