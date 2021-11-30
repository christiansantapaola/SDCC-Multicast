package internal

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"strings"
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
	nsAddr := os.Getenv("SDCC_NAME_SERVER_ADDRESS")
	fmt.Println(nsAddr)
	if nsAddr == "" {
		nsAddr = "127.0.0.1"
	}
	nsPort := os.Getenv("SDCC_NAME_SERVER_PORT")
	if nsPort == "" {
		nsPort = "2080"
	}
	userIP := os.Getenv("SDCC_HOST_IP")
	if userIP == "" {
		userIP = "127.0.0.1"
	}
	userPort := os.Getenv("SDCC_HOST_PORT")
	if userPort == "" {
		userPort = "2079"
	}
	verbose := os.Getenv("SDCC_VERBOSE")
	if verbose == "" {
		verbose = "false"
	}
	userPortInt, err := strconv.ParseInt(userPort, 10, 32)
	if err != nil {
		return err
	}
	var verboseBool bool = false
	if strings.ToLower(verbose) == "true" {
		verboseBool = true
	} else {
		verboseBool = false
	}
	cfg := Config{
		NameServerAddress: fmt.Sprintf("%s:%s", nsAddr, nsPort),
		UserId:            "",
		UserIp:            userIP,
		UserPort:          int(userPortInt),
		Secure:            false,
		Verbose:           verboseBool,
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
