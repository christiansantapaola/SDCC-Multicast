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
	"time"
)

func createUser(ctx context.Context, cl nameservice.NameServiceClient, ip string, port int) *nameservice.User {
	user, err := client.CreateUser(ctx, cl, ip, int32(port))
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	return user
}

func createGroup(ctx context.Context, cl nameservice.NameServiceClient, group, id, ip string, port int) {
	groupInfo, err := client.CreateGroup(ctx, cl, group, &nameservice.User{Id: id, Ip: ip, Port: int32(port)})
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Printf("%s\n", "Group Info")
	fmt.Printf("%s: %s\n", "Group Name", groupInfo.Name)
	fmt.Printf("%s: %s\n", "Group Type", groupInfo.GetType().String())
}

func joinGroup(ctx context.Context, cl nameservice.NameServiceClient, group, id, ip string, port int) {
	groupInfo, err := client.JoinGroup(ctx, cl, group, &nameservice.User{Id: id, Ip: ip, Port: int32(port)})
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
	fmt.Printf("%s: %s\n", "user ID", user.Id)
	fmt.Printf("%s: %s\n", "address", user.Ip)
	fmt.Printf("%s: %d\n", "port", user.Port)
}

func getAddresses(ctx context.Context, cl nameservice.NameServiceClient, group string) {
	users, err := client.GetAddressGroup(ctx, cl, group)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	for _, user := range users {
		fmt.Printf("%s\n", "User Info")
		fmt.Printf("%s: %s\n", "user ID", user.Id)
		fmt.Printf("%s: %s\n", "address", user.Ip)
		fmt.Printf("%s: %d\n", "port", user.Port)
	}
}

func GetGRPCConn(serverAddress string, secure bool) (*nameservice.NameServiceClient, error) {
	var opts []grpc.DialOption
	if secure {
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
	flag.Parse()
	if len(os.Args) < 2 {
		fmt.Println(fmt.Errorf("%s [create-user|create-group|join-group|get-address|get-addresses]", os.Args[0]))
		os.Exit(1)
	}
	switch os.Args[1] {
	case "create-user":
		createUserCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		ip := createUserCmd.String("ip", "127.0.0.1", "ip to register")
		port := createUserCmd.Int("port", 2079, "port to register")
		serverAddress := createUserCmd.String("server-address", "127.0.0.1:2080", "the address of the server")
		secure := createUserCmd.Bool("secure", true, "use secure connection if false")
		timeout := createUserCmd.Duration("timeout", 1*time.Second, "timeout for the connection.")
		help := createUserCmd.Bool("help", false, "print this help message.")
		out := createUserCmd.String("out", "sdccLC_config.yaml", "output")
		ctx, _ := context.WithTimeout(context.Background(), *timeout)
		createUserCmd.Parse(os.Args[2:])
		if *help {
			createUserCmd.PrintDefaults()
			os.Exit(0)
		}
		cl, err := GetGRPCConn(*serverAddress, *secure)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintln(err))
			os.Exit(1)
		}
		user := createUser(ctx, *cl, *ip, *port)
		cfg, err := internal.ReadCfg(*out)
		if err != nil {
			err1 := internal.GenDefaultCfg(*out)
			if err1 != nil {
				log.Fatalln(err1)
			}
			cfg, err1 = internal.ReadCfg(*out)
			if err1 != nil {
				log.Fatalln(err1)
			}
		}
		cfg.UserId = user.GetId()
		cfg.UserPort = int(user.GetPort())
		cfg.UserIp = user.GetIp()
		err = internal.Update(*out, cfg)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	case "create-group":
		createGroupCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		group := createGroupCmd.String("group", "", "name of the group")
		id := createGroupCmd.String("id", "", "id of the user")
		ip := createGroupCmd.String("ip", "127.0.0.1", "ip to register")
		port := createGroupCmd.Int("port", 2080, "port to register")
		serverAddress := createGroupCmd.String("server-address", "127.0.0.1:2080", "the address of the server")
		secure := createGroupCmd.Bool("secure", true, "use secure connection if false")
		timeout := createGroupCmd.Duration("timeout", 1*time.Second, "timeout for the connection.")
		help := createGroupCmd.Bool("help", false, "print this help message.")
		ctx, _ := context.WithTimeout(context.Background(), *timeout)
		createGroupCmd.Parse(os.Args[2:])
		if *help {
			createGroupCmd.PrintDefaults()
			os.Exit(0)
		}
		cl, err := GetGRPCConn(*serverAddress, *secure)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintln(err))
			os.Exit(1)
		}
		createGroupCmd.Parse(os.Args[2:])
		createGroup(ctx, *cl, *group, *id, *ip, *port)
		os.Exit(0)
	case "join-group":
		joinGroupCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		group := joinGroupCmd.String("group", "", "name of the group")
		id := joinGroupCmd.String("id", "", "id of the user")
		ip := joinGroupCmd.String("ip", "127.0.0.1", "ip to register")
		port := joinGroupCmd.Int("port", 2080, "port to register")
		serverAddress := joinGroupCmd.String("server-address", "127.0.0.1:2080", "the address of the server")
		secure := joinGroupCmd.Bool("secure", true, "use secure connection if false")
		timeout := joinGroupCmd.Duration("timeout", 1*time.Second, "timeout for the connection.")
		help := joinGroupCmd.Bool("help", false, "print this help message.")
		ctx, _ := context.WithTimeout(context.Background(), *timeout)
		joinGroupCmd.Parse(os.Args[2:])
		if *help {
			joinGroupCmd.PrintDefaults()
			os.Exit(0)
		}
		cl, err := GetGRPCConn(*serverAddress, *secure)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintln(err))
			os.Exit(1)
		}
		joinGroup(ctx, *cl, *group, *id, *ip, *port)
		os.Exit(0)
	case "get-address":
		getAddressCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		id := getAddressCmd.String("group", "", "group of the user")
		serverAddress := getAddressCmd.String("server-address", "127.0.0.1:2080", "the address of the server")
		secure := getAddressCmd.Bool("secure", true, "use secure connection if false")
		timeout := getAddressCmd.Duration("timeout", 1*time.Second, "timeout for the connection.")
		help := getAddressCmd.Bool("help", false, "print this help message.")
		ctx, _ := context.WithTimeout(context.Background(), *timeout)
		getAddressCmd.Parse(os.Args[2:])
		if *help {
			getAddressCmd.PrintDefaults()
			os.Exit(0)
		}
		cl, err := GetGRPCConn(*serverAddress, *secure)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintln(err))
			os.Exit(1)
		}
		getAddress(ctx, *cl, *id)
		os.Exit(0)
	case "get-addresses":
		getAddressCmd := flag.NewFlagSet("foo", flag.ExitOnError)
		group := getAddressCmd.String("group", "", "group to query")
		serverAddress := getAddressCmd.String("server-address", "127.0.0.1:2080", "the address of the server")
		secure := getAddressCmd.Bool("secure", true, "use secure connection if false")
		timeout := getAddressCmd.Duration("timeout", 1*time.Second, "timeout for the connection.")
		help := getAddressCmd.Bool("help", false, "print this help message.")
		ctx, _ := context.WithTimeout(context.Background(), *timeout)
		getAddressCmd.Parse(os.Args[2:])
		if *help {
			getAddressCmd.PrintDefaults()
			os.Exit(0)
		}
		cl, err := GetGRPCConn(*serverAddress, *secure)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintln(err))
			os.Exit(1)
		}
		getAddresses(ctx, *cl, *group)
		os.Exit(0)

	default:
		fmt.Println(fmt.Errorf("%s: unknown command", os.Args[1]))
		os.Exit(1)
	}
}
