package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"sdcc/internal"
	"sdcc/pkg/nameservice/client"
	"sdcc/pkg/nameservice/nameservice"
)

/*
	Questo eseguibile richiede i partecipanti ad un gruppo esistente al name server.
	Richiede la presenza di un file di configurazione generato dall'apposito eseguibile
	`sdcc_gen_config` configurato con i parametri di connessione.
*/

func GetGRPCConn(serverAddress string, secure bool) (*nameservice.NameServiceClient, error) {
	var opts []grpc.DialOption
	if !secure {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(serverAddress, opts...)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	cl := nameservice.NewNameServiceClient(conn)
	return &cl, nil
}

func getAddresses(ctx context.Context, cl nameservice.NameServiceClient, group string) []*nameservice.User {
	users, err := client.GetAddressGroup(ctx, cl, group)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return users
}

func main() {
	group := flag.String("group", "", "name of the group")
	config := flag.String("config", "sdcc_config.yaml", "path of the config")
	help := flag.Bool("help", false, "print this help message.")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		return
	}
	cfg, err := internal.ReadCfg(*config)
	if err != nil {
		log.Fatalln(err)
	}
	cl, err := GetGRPCConn(cfg.NameServerAddress, cfg.Secure)
	if err != nil {
		log.Fatalln(err)
	}
	users := getAddresses(context.Background(), *cl, *group)
	for _, user := range users {
		fmt.Printf("%s\n", "User Info")
		fmt.Printf("%s: %s\n", "user ID", user.Id)
		fmt.Printf("%s: %s\n", "address", user.Ip)
		fmt.Printf("%s: %d\n", "port", user.Port)
	}
	os.Exit(0)
}
