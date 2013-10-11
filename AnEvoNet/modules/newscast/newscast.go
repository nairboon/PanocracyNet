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
	"fmt"
	"github.com/samuel/go-thrift/thrift"
	ae "libanevonet"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	proto "newscast_protocol"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Newscast struct {
	mu    sync.Mutex
	State *proto.PeerState
	Con   *ae.AnEvoConnection
	DNA   Common.P2PDNA
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
		if len(n.State.NewsItems) < 1 {
			log.Printf("we have 0 peers in cache")
			// we have no peers at all!
			pl, err := n.Con.Bootstrap()
			if err == nil {
				for _, p := range pl {
					fmt.Printf("New peer %+v\n", p)
					cacheentry := &proto.CacheEntry{Agent: p, Time: &proto.Timestamp{Sec: time.Now().Unix()}}
					n.State.NewsItems = append(n.State.NewsItems, cacheentry)
				}
			}
			// still no peers
			if len(n.State.NewsItems) < 1 {
				time.Sleep(2000 * time.Millisecond)
			}
			continue
		}
		// select Peer
		peer := n.State.NewsItems[r.Intn(len(n.State.NewsItems))].Agent
		c, err := n.Con.GetPeerConnection(peer)
		if err != nil {
			log.Printf("could not get peer connectin :/")
			continue
		}
		pc := proto.NewscastClient{c}
		log.Printf("going to gossip...")
		recstate, err := pc.ExchangeState(n.State)
		if err != nil {
			panic(err)
		}
		n.UpdateState(recstate)
		time.Sleep(time.Duration(n.DNA["sleep"]) * time.Millisecond)
	}

}

func (n *Newscast) ExchangeState(state *proto.PeerState) (*proto.PeerState, error) {

	return n.State, nil
}

func (n *Newscast) PassiveThread(socket string) {

	rpc.RegisterName("Thrift", &proto.NewscastServer{n})

	ln, err := net.Listen("unix", socket)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("ERROR: %+v\n", err)
			continue
		}
		fmt.Printf("New connection %+v\n", conn)
		go rpc.ServeCodec(thrift.NewServerCodec(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false)))
	}

}

func main() {
	log.Printf("newscasting")

	nc := Newscast{}
	nc.Con = ae.NewModule("Newscast")
	socket := nc.Con.Register(proto.RootDNA, &nc.DNA)
	nc.State = &proto.PeerState{}

	_ = socket
	//go nc.PassiveThread(socket)
	log.Printf("starting active thread")
	_, _ = nc.Con.Rpc.Status()
	go nc.ActiveThread()
	/* connect to daemon
	   register protocol
	   name, dna
	   setup handlers

	 	for
		DNA change
		incoming msg
	   start out msg loop
	*/
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for _ = range c {
		break
	}

	fmt.Println("stopping newscast")
}
