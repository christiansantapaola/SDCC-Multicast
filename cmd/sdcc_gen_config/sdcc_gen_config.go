package main

import (
	"flag"
	"log"
	"os"
	"sdcc/internal"
)

func main() {
	config := flag.String("config", "sdcc_config.yaml", "path to config file")
	help := flag.Bool("help", false, "print this help message.")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	err := internal.GenDefaultCfg(*config)
	if err != nil {
		log.Fatalln(err)
	}
}
