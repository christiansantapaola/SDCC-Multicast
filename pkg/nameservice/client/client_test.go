package client

import (
	"fmt"
	"google.golang.org/grpc"
	pb "sdcc/pkg/nameservice/nameservice"
	"testing"
)

func InitClient(ip string, port uint) (pb.NameServiceClient, error){
	dial, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewNameServiceClient(dial), nil

}

func TestCreateUser(t *testing.T) {
	client, err := InitClient("localhost", 10000)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	user, err := CreateUser(client, "localhost", 50051)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	fmt.Printf("%s, %d, %d", user.Ip, user.Port, user.Id)
}

func TestGetAddress(t *testing.T) {
	client, err := InitClient("localhost", 10000)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	user, err := GetAddress(client, 0)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	t.Logf("%s, %d, %d", user.Ip, user.Port, user.Id)
}

func TestJoinGroup(t *testing.T) {
	client, err := InitClient("localhost", 10000)
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	group, err := JoinGroup(client, "g1", &pb.User{Id: 1, Ip: "localhost", Port: 50051})
	if err != nil {
		t.Fatalf("[ERROR]: %v", err)
	}
	t.Logf("%s", group.Name)

}