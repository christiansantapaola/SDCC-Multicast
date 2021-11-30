package main

import (
	"flag"
	"fmt"
	pb "github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/nameservice"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"os"
)

/*
	Questo eseguibile instanzia un ending point del name service.
	Richiede un istanza attiva del datastore etcd, i cui endpoint devono essere specificati nel file di configurazione.
	La comunicazione avviene tramite grpc e protobuf.
*/

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 2080, "The server port")
	config   = flag.String("config", "sdcc_name_server_config.yml", "path to the config file")
	help     = flag.Bool("help", false, "print this help message.")
)

func main() {
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		fmt.Println("SDCC_NAME_SERVER_ETCD: list of etcd endpoint in format\"addr1:port1;addr2:port2\" ")
		os.Exit(0)
	}
	cfg, err := server.ReadCfg(*config)
	if err != nil {
		log.Printf("failed to read config: %v", err)
		log.Printf("generating default config!")
		err := server.GenDefaultCfg(*config)
		if err != nil {
			log.Fatalf("Can't generate default config: %v", err)
		}
		cfg, err = server.ReadCfg(*config)
		if err != nil {
			log.Fatalf("INTERNAL ERROR: can't ReadCfg after generating it")
		}
	}
	log.Printf("Read config at '%s' succesfully", *config)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	log.Printf("opening new grpc server on port %d", *port)
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterNameServiceServer(grpcServer, server.NewNamingService(cfg))
	log.Printf("start Listening")
	grpcServer.Serve(lis)
}
