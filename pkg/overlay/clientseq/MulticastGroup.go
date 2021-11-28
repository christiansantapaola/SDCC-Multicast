package clientseq

import (
	"fmt"
	"log"
	"net"
	pb "sdcc/pkg/nameservice/nameservice"
	"sort"
)

type Users []*pb.User

func (users Users) Len() int           { return len(users) }
func (users Users) Swap(i, j int)      { users[i], users[j] = users[j], users[i] }
func (users Users) Less(i, j int) bool { return users[i].GetId() < users[j].GetId() }

type MulticastGroup struct {
	selfIndex int
	users     Users
}

func NewGroup(self string, users []*pb.User) *MulticastGroup {
	selfIndex := -1
	mgroup := MulticastGroup{
		users:     users,
		selfIndex: selfIndex,
	}
	sort.Sort(mgroup.users)
	for index, user := range users {
		if user.GetId() == self {
			selfIndex = index
		}
	}
	if selfIndex == -1 {
		log.Fatalf("[FATAL ERROR] not part of the group, register first\n")
	}
	mgroup.selfIndex = selfIndex
	return &mgroup
}

func (group *MulticastGroup) GetRank(ID string) (int, error) {
	for index, user := range group.users {
		if user.GetId() == ID {
			return index, nil
		}
	}
	return -1, fmt.Errorf("'%s' not found", ID)
}

func (group *MulticastGroup) GetMyRank() (int, error) {
	return group.GetRank(group.GetMyID())
}

func (group *MulticastGroup) GetID(rank int) (string, error) {
	if rank >= len(group.users) {
		return "", nil
	}
	return group.users[rank].GetId(), nil
}

func (group *MulticastGroup) GetMyID() string {
	return group.users[group.selfIndex].GetId()
}

func (group *MulticastGroup) GetSequencerServices() (net.Addr, error) {
	if len(group.users) < 1 {
		return nil, fmt.Errorf("no users")
	}
	seq := group.users[0]
	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", seq.Ip, seq.Port))
}

func (group *MulticastGroup) GetUsersInfo() ([]net.Addr, []string, []int, error) {
	size := len(group.users)
	var groupAddrs []net.Addr = make([]net.Addr, size)
	var groupIDs []string = make([]string, size)
	var groupPorts []int = make([]int, size)
	for index, user := range group.users {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", user.Ip, user.Port))
		if err != nil {
			return nil, nil, nil, err
		}
		groupAddrs[index] = addr
		groupIDs[index] = user.GetId()
		groupPorts[index] = int(user.GetPort())
	}
	return groupAddrs, groupIDs, groupPorts, nil
}