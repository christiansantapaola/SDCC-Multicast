package clientlc

import (
	"fmt"
	pb "github.com/christiansantapaola/SDCC-Multicast/pkg/nameservice/nameservice"
	"log"
	"net"
	"sort"
)

/*
	MulticastGroup
	Struttura che si ordina di dare ordine al gruppo.
	Recupera il numero di utenti nel gruppo, il nostro rank al suo interno, il nostro id.
	Il rank viene effettuato ordinado gli ID in ordine lessicografico.
*/

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

func (group *MulticastGroup) GetShortID() string {
	return group.GetMyID()[:6]
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
