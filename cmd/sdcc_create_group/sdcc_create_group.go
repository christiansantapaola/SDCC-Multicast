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
	Questo eseguibile registra un nuovo gruppo nell'istanza del name server.
	Per creare un gruppo bisogna essersi registrati come utenti al name server in precedenza.
	Il gruppo è identificato da un nome, il nome deve essere univoco.
	Il programma richiede la presenza di un config file generato dall'apposito eseguibile ´sdcc_gen_config´ e
	configurato con i propri parametri di connessione.
*/

func createGroup(ctx context.Context, cl nameservice.NameServiceClient, group, id, ip string, port int) *nameservice.Group {
	groupInfo, err := client.CreateGroup(ctx, cl, group, &nameservice.User{Id: id, Ip: ip, Port: int32(port)})
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return groupInfo
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
	group := flag.String("group", "", "name of the group")
	config := flag.String("config", "sdcc_config.yaml", "path of the config")
	help := flag.Bool("help", false, "print this help message.")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	cfg, err := internal.ReadCfg(*config)
	if err != nil {
		log.Fatalf("config at '%s' not found!\n", *config)
	}
	ctx, _ := context.WithTimeout(context.Background(), cfg.Timeout)
	cl, err := GetGRPCConn(cfg.NameServerAddress, cfg.Secure)
	if err != nil {
		log.Fatalln(err)
	}
	groupInfo := createGroup(ctx, *cl, *group, cfg.UserId, cfg.UserIp, cfg.UserPort)
	fmt.Printf("%s\n", "Group Info")
	fmt.Printf("%s: %s\n", "Group Name", groupInfo.Name)
	os.Exit(0)
}
