package client

import (
	"context"
	"fmt"
	pb "github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/nameservice"
	"google.golang.org/grpc"
	"testing"
)

func InitClient(ip string, port uint) (pb.NameServiceClient, error) {
	dial, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewNameServiceClient(dial), nil

}

func TestCreateUser(t *testing.T) {
	client, err := InitClient("localhost", 2180)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	user, err := CreateUser(context.Background(), client, "localhost", 50051)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	fmt.Printf("%s, %d, %s", user.Ip, user.Port, user.Id)
}

func TestGetAddress(t *testing.T) {
	client, err := InitClient("localhost", 2180)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	user, err := GetAddress(context.Background(), client, "abc")
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	t.Logf("%s, %d, %s", user.Ip, user.Port, user.Id)
}

func TestJoinGroup(t *testing.T) {
	client, err := InitClient("localhost", 2180)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	group, err := JoinGroup(context.Background(), client, "g1", &pb.User{Id: "abc", Ip: "localhost", Port: 50051})
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	t.Logf("%s", group.Name)

}
