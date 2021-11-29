package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/christiansantapaola/SDCC-Multicast/internal"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/client"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/nameservice"
	"google.golang.org/grpc"
	"log"
	"os"
)

/*
	Questo eseguibile registra un nuovo utente ad un istanza attiva del name server.
	Richiede la presenza di un file di configurazione generato dall'apposito eseguibile
	`sdcc_gen_config` configurato con i parametri di connessione.
	Questo eseguibile sovvrascrive il file di configurazione immettendo le informazioni di
	registrazione, quali l'ID all'interno della rete di overlay.
*/

func createUser(ctx context.Context, cl nameservice.NameServiceClient, ip string, port int) *nameservice.User {
	user, err := client.CreateUser(ctx, cl, ip, int32(port))
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return user
}

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

func main() {
	help := flag.Bool("help", false, "print this help message.")
	config := flag.String("config", "sdcc_config.yaml", "path to config file")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *config == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	cfg, err := internal.ReadCfg(*config)
	if err != nil {
		log.Fatalln(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), cfg.Timeout)
	cl, err := GetGRPCConn(cfg.NameServerAddress, cfg.Secure)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintln(err))
		os.Exit(1)
	}
	user := createUser(ctx, *cl, cfg.UserIp, cfg.UserPort)
	cfg.UserId = user.GetId()
	cfg.UserPort = int(user.GetPort())
	cfg.UserIp = user.GetIp()
	err = internal.Update(*config, cfg)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
