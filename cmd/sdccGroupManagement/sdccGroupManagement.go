package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"sdcc/pkg/nameservice/client"
	"sdcc/pkg/nameservice/nameservice"
	"time"
)

var (
	serverAddress = flag.String("server-address", "127.0.0.1:2080", "the address of the server")
	secure        = flag.Bool("secure", true, "use secure connection if false")
	timeout       = flag.Duration("timeout", 1 * time.Second, "timeout for the connection.")
)

func createUser(ctx context.Context, cl nameservice.NameServiceClient, ip string, port uint32) {
	user, err := client.CreateUser(ctx, cl, ip, port)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Printf("%s\n", "User Info")
	fmt.Printf("%s: %s", "user ID", user.Id)
	fmt.Printf("%s: %s", "address", user.Ip)
	fmt.Printf("%s: %d", "port", user.Port)
}

func createGroup(ctx context.Context, cl nameservice.NameServiceClient, group, id ,ip string, port uint32) {
	groupInfo, err := client.CreateGroup(ctx, cl, group, &nameservice.User{Id: id, Ip: ip, Port: port})
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Printf("%s\n", "Group Info")
	fmt.Printf("%s: %s\n", "Group Name", groupInfo.Name)
	fmt.Printf("%s: %s\n", "Group Type", groupInfo.GetType().String())
}

func joinGroup(ctx context.Context, cl nameservice.NameServiceClient, group, id ,ip string, port uint32) {
	groupInfo, err := client.JoinGroup(ctx, cl, group, &nameservice.User{Id: id, Ip: ip, Port: port})
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Printf("%s\n", "Group Info")
	fmt.Printf("%s: %s\n", "Group Name", groupInfo.Name)
	fmt.Printf("%s: %s\n", "Group Type", groupInfo.GetType().String())
}

func getAddress(ctx context.Context, cl nameservice.NameServiceClient, id string) {
	user, err := client.GetAddress(ctx, cl, id)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Printf("%s\n", "User Info")
	fmt.Printf("%s: %s", "user ID", user.Id)
	fmt.Printf("%s: %s", "address", user.Ip)
	fmt.Printf("%s: %d", "port", user.Port)
}


func main()  {
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println(fmt.Errorf("%s [create-user|create-group|join-group|get-address]", os.Args[0]))
		flag.PrintDefaults()
		os.Exit(1)
	}
	var opts []grpc.DialOption
	if *secure {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddress, opts...)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	ctx, _ := context.WithTimeout(context.Background(), *timeout)
	cl := nameservice.NewNameServiceClient(conn)
	switch os.Args[1] {
	case "create-user":
		createUserCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		ip := createUserCmd.String("ip", "127.0.0.1", "ip to register")
		port := createUserCmd.Uint("port", 2079, "port to register")
		createUserCmd.Parse(os.Args[2:])
		createUser(ctx, cl, *ip, uint32(*port))
		os.Exit(0)
	case "create-group":
		createGroupCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		group := createGroupCmd.String("group", "", "name of the group")
		id := createGroupCmd.String("id", "", "id of the user")
		ip := createGroupCmd.String("ip", "127.0.0.1", "ip to register")
		port := createGroupCmd.Uint("port", 2080, "port to register")
		createGroupCmd.Parse(os.Args[2:])
		createGroup(ctx, cl, *group, *id, *ip, uint32(*port))
		os.Exit(0)
	case "join-group":
		joinGroupCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		group := joinGroupCmd.String("group", "", "name of the group")
		id := joinGroupCmd.String("id", "", "id of the user")
		ip := joinGroupCmd.String("ip", "127.0.0.1", "ip to register")
		port := joinGroupCmd.Uint("port", 2080, "port to register")
		joinGroupCmd.Parse(os.Args[2:])
		createGroup(ctx, cl, *group, *id, *ip, uint32(*port))
		os.Exit(0)
	case "get-address":
		getAddressCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		id := getAddressCmd.String("id", "", "id of the user")
		getAddressCmd.Parse(os.Args[2:])
		getAddress(ctx, cl, *id)
		os.Exit(0)
	default:
		fmt.Println(fmt.Errorf("%s: unknown command", os.Args[1]))
		os.Exit(1)
	}
}
