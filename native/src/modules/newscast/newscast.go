package main

/* Newscast Algorithm according to "Tag-Based Cooperation in Peer-to-Peer Networks with Newscast" (2005)

ACTIVE THREAD
while(TRUE) do
wait(!t);
neighbor = SELECTPEER();
SENDSTATE(neighbor);
n.state = RECEIVESTATE();
my.state.UPDATE(n.state);


PASSIVE THREAD
while(TRUE) do
n.state = RECEIVESTATE();
SENDSTATE(n.state.sender);
my.state.UPDATE(n.state);
*/

import (
	//ae "github.com/nairboon/anevonet/lib"
	"Common"
	ae "libanevonet"
	"log"
	"math/rand"
	proto "newscast_protocol"
	"sync"
	"time"
)

type Newscast struct {
	mu    sync.Mutex
	State *proto.PeerState
	Con   *ae.AnEvoConnection
	DNA   Common.DNA
}

func (n *Newscast) UpdateState(newstate *proto.PeerState) proto.PeerState {
	n.mu.Lock()
	defer n.mu.Unlock()

	// check if we need more peers
	if len(n.State.NewsItems) < int(n.DNA["cachesize"]) {
		// we should
	}

	return *newstate
}

func (n *Newscast) ActiveThread() {
	r := rand.New(rand.NewSource(99))
	for n.Con.ContinueRunning(n.DNA) {

		// select Peer
		peer := n.State.NewsItems[r.Intn(len(n.State.NewsItems))].Agent
		pc := proto.NewscastClient{n.Con.GetPeerConnection(peer)}

		recstate, err := pc.ExchangeState(n.State)
		if err != nil {
			panic(err)
		}
		n.UpdateState(recstate)

		time.Sleep(time.Duration(n.DNA["sleep"]) * time.Millisecond)

	}

}

func (n *Newscast) PassiveThread() {

}

func main() {
	log.Printf("newscasting")

	nc := Newscast{}
	nc.Con = ae.NewConnection()
	nc.Con.Register("Newscast", proto.RootDNA, nc.DNA)
	nc.State = &proto.PeerState{}

	/* connect to daemon
	   register protocol
	   name, dna
	   setup handlers

	 	for
		DNA change
		incoming msg
	   start out msg loop
	*/

}
