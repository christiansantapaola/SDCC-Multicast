package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/christiansantapaola/SDCC-Multicast/internal"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/client"
	"github.com/christiansantapaola/SDCC-Multicast/pkg/overlay/clientlc"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

/*
	Utility di test del servizio multicast a clock scalare.
	Questa utility manda un numero prestabilito di messaggi alternati da un delay randomico dai 5 ai 10 secondi.
*/

func SendDelay(ctx context.Context, lc *clientlc.MiddlewareLC, message string, delay time.Duration) error {
	time.Sleep(delay)
	err := lc.Send(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

func SendWork(lc *clientlc.MiddlewareLC, n int, seed int64) {
	i := 0
	rand.Seed(seed)
	for j := 0; j < n; j++ {
		delay := time.Duration(5+rand.Intn(5)) * time.Second
		str := fmt.Sprintf("%s:%s:%d", "msg", lc.GetShortID(), i)
		//fmt.Printf("sending: %s\n", str)
		err := SendDelay(context.Background(), lc, str, delay)
		if err != nil {
			log.Println(err)
			// os.Exit(1)
		}
		//fmt.Printf("send: %s\n", str)
		i++
	}
}

func main() {
	helpFlag := flag.Bool("help", false, "Print this help message.")
	configPath := flag.String("config", "sdcc_config.yaml", "path to the config")
	groupName := flag.String("group", "", "the group to connect")
	logPath := flag.String("logPath", "log.txt", "path to log")
	numMsg := flag.Int("numMsg", 10, "number of message to send")
	flag.Parse()
	if *helpFlag || len(os.Args) < 2 {
		flag.PrintDefaults()
		os.Exit(0)
	}
	cfg, err := internal.ReadCfg(*configPath)
	if err != nil {
		err1 := internal.GenDefaultCfg(*configPath)
		if err1 != nil {
			log.Fatalf("[ERROR FATAL] can't create config: %v\n", err)
		}
		cfg, err1 = internal.ReadCfg(*configPath)
		if err1 != nil {
			log.Fatalf("[ERROR FATAL] can't create config: %v\n", err)
		}
	}

	nameServerAddr, err := net.ResolveTCPAddr("tcp", cfg.NameServerAddress)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	var opts []grpc.DialOption
	var sopt []grpc.ServerOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithInsecure())
	}
	ns, err := client.NewClient(nameServerAddr, opts)
	if err != nil {
		log.Fatalln(err)
	}
	middleware, err := clientlc.NewMiddlewareLC(cfg.UserId, *groupName, *logPath, cfg.UserPort, ns, cfg.Verbose, false, opts, sopt)
	if err != nil {
		log.Fatalln("%v\n", err)
	}
	id := middleware.GetGroupID()
	rank, _ := middleware.GetRank()
	fmt.Printf("MyID: %s\nmyRank: %d\nShortID: %s\n", id, rank, middleware.GetShortID())
	//err = middleware.WaitToStart(context.Background())
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	n := *numMsg
	go SendWork(middleware, n, int64(4096*rank))
	for i := 0; i < n*middleware.GetGroupSize(); i++ {
		msg, err := middleware.Recv(context.Background())
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("%s: %s\n", "[MESSAGE] Data", msg)
	}
	middleware.Stop()
}
