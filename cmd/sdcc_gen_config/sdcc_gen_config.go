package main

import (
	"flag"
	"fmt"
	"github.com/christiansantapaola/SDCC-Multicast/internal"
	"log"
	"os"
)

/*
	Questo eseguibile genera uno scheletro di configurazione.
	Il file di configurazione è in formato yaml
*/

func main() {
	config := flag.String("config", "sdcc_config.yaml", "path to config file")
	help := flag.Bool("help", false, "print this help message.")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		fmt.Println("VARIABLE ENVIRONEMNT LIST:\n" +
			"SDCC_NAME_SERVER_ADDRESS: ADDRESS OF THE NAME SERVICE\n" +
			"SDCC_NAME_SERVER_PORT: PORT OF THE NAME SERVICE\n" +
			"SDCC_HOST_IP: REACHABLE IP OF THE HOST\n" +
			"SDCC_HOST_PORT: OPEN PORT FOR THE HOST\n" +
			"SDCC_VERBOSE: RUN IN DEBUG MODE")
		os.Exit(0)
	}
	err := internal.GenDefaultCfg(*config)
	if err != nil {
		log.Fatalln(err)
	}
}
