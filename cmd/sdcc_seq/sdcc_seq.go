package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"os"
	"sdcc/internal"
	"sdcc/pkg/nameservice/client"
	"sdcc/pkg/overlay/clientseq"
	"time"
)

func SendDelay(ctx context.Context, seq *clientseq.MiddlewareSeq, message string, delay time.Duration) error {
	time.Sleep(delay)
	err := seq.Send(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

func SendWork(lc *clientseq.MiddlewareSeq, n int, self string) {
	i := 0
	for j := 0; j < n; j++ {
		delay := time.Duration(time.Duration(5+rand.Intn(5)) * time.Second)
		str := fmt.Sprintf("%s:%d", self[:5], i)
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
		log.Println(err)
	}
	middleware, err := clientseq.NewMiddlewareSeq(cfg.UserId, *groupName, *logPath, cfg.UserPort, ns, cfg.Verbose, opts, sopt)
	if err != nil {
		log.Println("%v\n", err)
	}
	id := middleware.GetGroupID()
	rank, _ := middleware.GetGroupRank()
	fmt.Printf("ID: '%s'\nRank: %d\nAm I the sequencer: %t\n\n", id, rank, middleware.AmITheSequencer())
	n := 10
	go SendWork(middleware, n, cfg.UserId)
	for i := 0; i < n*middleware.GetGroupSize(); i++ {
		msg, err := middleware.Recv(context.Background())
		if err != nil {
			log.Printf("[main]: %v\n", err)
		}
		fmt.Printf("%s: %s\n", "[MESSAGE] Data: ", msg)
	}
	middleware.Stop()
	time.Sleep(11 * time.Second)
}
