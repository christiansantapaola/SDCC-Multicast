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
	"sdcc/pkg/nameservice/client"
	"sdcc/pkg/overlay/clientlc"
	"time"
)

func SendDelay(ctx context.Context, lc *clientlc.MiddlewareLC, message string, delay time.Duration) error {
	time.Sleep(delay)
	err := lc.Send(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

func SendWork(lc *clientlc.MiddlewareLC, self string) {
	i := 0
	for {
		delay := time.Duration(time.Duration(rand.Intn(5)) * time.Second)
		err := SendDelay(context.Background(), lc, fmt.Sprintf("%s:%d", self, i), delay)
		if err != nil {
			os.Exit(1)
		}
	}
}

func main() {
	helpFlag := flag.Bool("help", false, "Print this help message.")
	nameServer := flag.String("server", "127.0.0.1:2080", "the address of the naming server")
	groupName := flag.String("group", "", "the group to connect")
	secure := flag.Bool("secure", true, "use non secure connection.")
	id := flag.String("id", "", "personal id")
	flag.Parse()
	fmt.Println(len(os.Args))
	if *helpFlag || len(os.Args) < 2 {
		flag.PrintDefaults()
		os.Exit(0)
	}
	nameServerAddr, err := net.ResolveTCPAddr("tcp", *nameServer)
	if err != nil {
		os.Exit(1)
	}
	var opts []grpc.DialOption
	if *secure {
		opts = append(opts, grpc.WithInsecure())
	}
	ns, err := client.NewClient(nameServerAddr, opts)
	if err != nil {
		return
	}
	group, err := client.GetAddressGroup(context.Background(), *ns, *groupName)
	if err != nil {
		return
	}
	var groupAddrs []net.Addr
	for _, user := range group {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", user.Ip, user.Port))
		if err != nil {
			log.Println(err)
			continue
		}
		groupAddrs = append(groupAddrs, addr)
	}
	middleware, err := clientlc.NewMiddlewareLC(*id, groupAddrs, opts)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	go SendWork(middleware, *id)
	for {
		msg := middleware.Recv(context.Background())
		fmt.Println("Message")
		fmt.Printf("%s: %s\n", "SOURCE", msg.GetSrc())
		fmt.Printf("%s: %d\n", "CLOCK", msg.GetClock())
		fmt.Printf("%s: %s\n", "ID", msg.GetId())
		fmt.Printf("%s: %s\n", "DATA", msg.GetData())
	}
}
