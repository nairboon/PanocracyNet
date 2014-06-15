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
	//"io/ioutil"
	//"errors"
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
	We    *proto.CacheEntry
	ID    *Common.Peer
}

func AddToPeerState(e *proto.CacheEntry, ps *proto.PeerState) bool {
	// do we already have this one?
	have := false
	log.Printf("trying to merge states %v", e)

	for _, c := range ps.NewsItems {
		if c.Agent.ID == e.Agent.ID {
			have = true
		}
	}
	if !have {
		ps.NewsItems = append(ps.NewsItems, e)
		log.Printf("we add peer: %s", e.Agent.ID)

	}
	return !have
}
func (n *Newscast) UpdateState(newstate *proto.PeerState) proto.PeerState {
	n.mu.Lock()
	defer n.mu.Unlock()

	log.Printf("Newscast updateState from %s with: %v", newstate.Source.ID, newstate.NewsItems)

	// check if we need more peers
	if len(n.State.NewsItems) < int(n.DNA["cachesize"]) {
		log.Printf("we should add some peers %d < %d", len(n.State.NewsItems), int(n.DNA["cachesize"]))

		// we should add some items from newstate
		res := false
		for _, ni := range newstate.NewsItems {
			res = AddToPeerState(ni, n.State)
			if len(n.State.NewsItems) >= int(n.DNA["cachesize"]) {
				break
			}

			if res {
				log.Printf("We added a peer!")
			}
		}
		if !res {
			log.Printf("No peer matched!?!?")

		}
	} else {
		log.Printf("we have enough peers (%d) already: %d", len(n.State.NewsItems), int(n.DNA["cachesize"]))

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
			time.Sleep(time.Duration(n.DNA["sleep"]) * time.Millisecond * 500)
			continue
		}
		pc := proto.NewscastClient{c}
		log.Printf("going to gossip...")

		timeout := make(chan bool, 1)
		ch := make(chan *proto.PeerState, 1)

		go func() {
			time.Sleep(5 * time.Second)
			timeout <- true
		}()
		go func() {
			recstate, _ := pc.ExchangeState(n.State)
			ch <- recstate
		}()
		select {
		case recstate := <-ch:
			// a read from ch has occurred

			log.Printf("updating state... %+v", recstate)

			/*if err != nil {
				panic(err)
			}*/
			n.UpdateState(recstate)
			time.Sleep(time.Duration(n.DNA["sleep"]) * time.Millisecond)

		case <-timeout:
			// the read from ch has timed out
			log.Printf("timed out")
			continue

		}

	}

}

func (n *Newscast) ExchangeState(state *proto.PeerState) (*proto.PeerState, error) {

	// adding ourselve to the state
	ns := append(n.State.NewsItems, n.We)

	log.Printf("%s called %+v ExchangeState %v", state.Source.ID, n.We, state.NewsItems)
	res := &proto.PeerState{Source: n.ID, NewsItems: n.State.NewsItems}

	a := ns
	b := n.State.NewsItems
	log.Printf("RES: %#v\n Difference: %+v ---- %+v", res, a, b)
	res.NewsItems = ns

	n.UpdateState(state)

	//res.Source.ID = "ABCDEFGHIJKLMNM"

	log.Printf("RES222: %+v", res.Source)
	res.NewsItems = nil
	return res, nil

	res = &proto.PeerState{Source: n.ID, NewsItems: ns}
	log.Printf("we (%s) respond: %v", n.ID.ID, res)
	return res, nil
}

func (n *Newscast) PassiveThread(socket string) {

	rpc.RegisterName("Thrift", &proto.NewscastServer{n})

	ln, err := net.Listen("tcp", socket)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("ERROR: %+v\n", err)
			continue
		}
		fmt.Printf("New connection in Passive Thread %+v\n", conn)
		//var re []byte
		//i, _ := conn.Read(re)
		//re, _ := ioutil.ReadAll(conn)
		go rpc.ServeCodec(thrift.NewServerCodec(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false)))
		//err = rpc.ServeRequest(thrift.NewServerCodec(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false)))

		//conn.Close()
		//fmt.Printf("Done processing with RPC ---------%v----!!!!!!\n", err)

	}

}

func main() {
	log.Printf("newscasting")

	nc := Newscast{}
	nc.Con, _ = ae.NewModule("Newscast")
	res, dna, _ := nc.Con.Register(proto.RootDNA)

	nc.DNA = dna

	socket := res.Socket
	nc.State = &proto.PeerState{Source: res.ID}

	nc.ID = res.ID
	nc.We = &proto.CacheEntry{Agent: nc.ID, Time: &proto.Timestamp{Sec: 99999}}
	_ = socket

	log.Printf("Initialized module our DNA: %#v", nc.DNA)

	log.Printf("starting passive thread on %s", socket)

	go nc.PassiveThread(socket)
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
