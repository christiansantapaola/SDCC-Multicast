package main

import (
	"flag"
	"fmt"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/server"
	"log"
	"os"
)

/*
	Questo eseguibile genera uno scheletro di configurazione.
	Il file di configurazione è in formato yaml
*/

func main() {
	config := flag.String("config", "sdcc_name_server_config.yaml", "path to config file")
	help := flag.Bool("help", false, "print this help message.")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		fmt.Println("SDCC_NAME_SERVER_ETCD: list of etcd endpoint in format\"addr1:port1;addr2:port2\" ")
		os.Exit(0)
	}
	err := server.GenDefaultCfg(*config)
	if err != nil {
		log.Fatalln(err)
	}
}
