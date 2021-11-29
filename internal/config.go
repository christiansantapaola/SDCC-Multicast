package internal

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

/*
	Qui è contenuto il codice che gestisce il file di configurazione dei client:
	NameServerAddress: l'indirizzo in cui il name server è in ascolto
	UserID: Id di un utente, rilasciato dal name server
	UserIP: IP dove il server grpc utente è in ascolto
	UserPort: Porta dove il server grpc utente è in ascolto
	Secure: flag se utilizzare tls o no per grpc
	Verbose: flag se far girare il client in modalita Verbose o meno
	Timeout: timeout per le connessione al name server.
*/

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
