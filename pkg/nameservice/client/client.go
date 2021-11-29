package client

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"io"
	"net"
	pb "sdcc/pkg/nameservice/nameservice"
)

/*
	Qui sono contenuti dei wrapper sopra le funzioni generate da grpc per il client.
*/

func NewClient(serverAddr net.Addr, opts []grpc.DialOption) (*pb.NameServiceClient, error) {
	conn, err := grpc.Dial(serverAddr.String(), opts...)
	if err != nil {
		return nil, err
	}
	client := pb.NewNameServiceClient(conn)
	return &client, nil
}

func CreateUser(ctx context.Context, client pb.NameServiceClient, ip string, port int32) (*pb.User, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	newUser := pb.NewUser{Ip: ip, Port: port}
	return client.CreateUser(ctx, &newUser)
}

func CreateGroup(ctx context.Context, client pb.NameServiceClient, group string, user *pb.User) (*pb.Group, error) {
	return client.CreateGroup(ctx, &pb.NewGroup{Name: group, Creator: user})
}

func JoinGroup(ctx context.Context, client pb.NameServiceClient, group string, user *pb.User) (*pb.Group, error) {
	return client.JoinGroup(ctx, &pb.JoinRequest{Group: group, User: user})
}

func GetAddress(ctx context.Context, client pb.NameServiceClient, userID string) (*pb.User, error) {
	return client.GetAddress(ctx, &pb.UserId{Id: userID})
}

func GetAddressGroup(ctx context.Context, client pb.NameServiceClient, group string) ([]*pb.User, error) {
	var users []*pb.User
	addresses, err := client.GetGroupAddresses(ctx, &pb.Group{Name: group})
	if err != nil {
		return nil, err
	}
	for {
		user, err := addresses.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
