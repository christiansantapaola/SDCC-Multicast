package server

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	pb "sdcc/pkg/nameservice/nameservice"
	"time"
)

type NamingServiceServer struct {
	pb.UnimplementedNameServiceServer
	DialTimeout time.Duration
	Endpoints   []string
}

func NewNamingService(cfg *Config) *NamingServiceServer {
	return &NamingServiceServer{DialTimeout: cfg.Etcd.DialTimeout, Endpoints: cfg.Etcd.Endpoints}
}

func genID(ip string, port int) string {
	h := sha256.New()
	idstr := fmt.Sprintf("%s:%d", ip, port)
	h.Write([]byte(idstr))
	bh := h.Sum(nil)
	return fmt.Sprintf("%x", bh)
}

func getUserID(id string) string {
	return "user/" + id
}

func getGroupKey(group string) string {
	return "group/" + group
}

func getGroupUserKey(group string, id string) string {
	return "group/" + group + "/" + id
}

func (s NamingServiceServer) CreateUser(ctx context.Context, newUser *pb.NewUser) (*pb.User, error) {
	cli, err := clientv3.New(clientv3.Config{DialTimeout: s.DialTimeout, Endpoints: s.Endpoints})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	userId := genID(newUser.Ip, int(newUser.Port))
	userKey := getUserID(userId)
	user := pb.User{Id: userId, Ip: newUser.Ip, Port: newUser.Port}
	data, err := proto.Marshal(&user)
	_, err = cli.KV.Put(ctx, userKey, string(data))
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s NamingServiceServer) CreateGroup(ctx context.Context, newGroup *pb.NewGroup) (*pb.Group, error) {
	cli, err := clientv3.New(clientv3.Config{DialTimeout: s.DialTimeout, Endpoints: s.Endpoints})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	groupKey := getGroupKey(newGroup.GetName())
	groupValue := getUserID(newGroup.Creator.GetId())
	_, err = cli.KV.Put(ctx, groupKey, groupValue)
	if err != nil {
		return nil, err
	}
	group := pb.Group{Name: newGroup.Name}
	return &group, nil
}

func (s NamingServiceServer) GetAddress(ctx context.Context, userID *pb.UserId) (*pb.User, error) {
	cli, err := clientv3.New(clientv3.Config{DialTimeout: s.DialTimeout, Endpoints: s.Endpoints})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	userKey := getUserID(userID.Id)
	get, err := cli.KV.Get(ctx, userKey)
	if len(get.Kvs) < 1 {
		return nil, errors.New("invalid ID")
	}
	var user pb.User
	err = proto.Unmarshal(get.Kvs[0].Value, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s NamingServiceServer) JoinGroup(ctx context.Context, group *pb.JoinRequest) (*pb.Group, error) {
	cli, err := clientv3.New(clientv3.Config{DialTimeout: s.DialTimeout, Endpoints: s.Endpoints})
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	key := getGroupUserKey(group.Group, group.User.Id)
	value, err := proto.Marshal(group.User)
	if err != nil {
		return nil, err
	}
	_, err = cli.KV.Put(ctx, key, string(value))
	if err != nil {
		return nil, err
	}
	return &pb.Group{Name: group.Group}, nil
}

func (s NamingServiceServer) GetGroupAddresses(group *pb.Group, stream pb.NameService_GetGroupAddressesServer) error {
	cli, err := clientv3.New(clientv3.Config{DialTimeout: s.DialTimeout, Endpoints: s.Endpoints})
	if err != nil {
		return err
	}
	defer cli.Close()
	groupKey := getGroupKey(group.Name) + "/"
	ctx, _ := context.WithTimeout(context.Background(), s.DialTimeout)
	get, err := cli.KV.Get(ctx, groupKey, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	for _, item := range get.Kvs {
		var user pb.User
		err = proto.Unmarshal(item.Value, &user)
		err := stream.Send(&user)
		if err != nil {
			return err
		}
	}
	return nil
}
