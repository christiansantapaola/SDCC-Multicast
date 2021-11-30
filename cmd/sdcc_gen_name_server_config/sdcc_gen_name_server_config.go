package main

import (
	"flag"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/server"
	"log"
	"os"
	"time"
)

/*
	Questo eseguibile genera uno scheletro di configurazione.
	Il file di configurazione è in formato yaml
*/

func main() {
	config := flag.String("config", "sdcc_name_server_config.yaml", "path to config file")
	endpoint := flag.String("etcd", "127.0.0.1:2079", "endpoint for etcd cluster")
	timeout := flag.Duration("timeout", 5*time.Second, "timeout for dial")
	help := flag.Bool("help", false, "print this help message.")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	err := server.GenDefaultCfg(*config, *timeout, *endpoint)
	if err != nil {
		log.Fatalln(err)
	}
}
