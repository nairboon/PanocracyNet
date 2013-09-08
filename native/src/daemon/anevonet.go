package main

/*
AnEvoNet daemon

responsible for connecting with remote peers and exchanging data, broker for modules

1. listen on zeromq for modules
1.1 modules can register
1.2 modules can shutdown

2. broker messages for modules
2.1 modules can request to talk with a remotepeer
2.1.2 daemon connects with remotepeer
2.1.3 daemon creates local socket designated to communicate with this peer
2.1.4 daemon sends socket to module

2.2 modules listen for remotepeers
2.2.1 daemon registers the socket where the module is listening
2.2.2 daemon pipes data from remotepeers to that socket


3. connects to remote peers
3.1 TODO: NATS
3.2 TODO: firewalls


4. evolves modules and stratgies
4.1 keeps statistics for every module
4.2 randomly mutates dna
4.3 exchanges dna with other peers
4.4 adapts strategy
4.4.1 if new strategy does not perform better, revert


5. out-of-time delivery of messages
5.1 store failed messages and retransmit on peer-up



#####
stored int daemon db

1. module statistics
2. known peers
2.1 lookup table of URIs for peers
3. out-of-connection delivery for messages
4. peer statistics i.e: who delivers out of connections msgs

*/
import (
	log "github.com/golang/glog"
	//flag //"github.com/ogier/pflag"
	"Common"
	irpc "anevonet_rpc"
	"errors"
	"flag"
	"fmt"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
	zmq3 "github.com/pebbe/zmq3"
	"github.com/samuel/go-thrift/thrift"
	zmq "libzmqthrift"

	"net/rpc"
	"os"
	"os/signal"
	"path/filepath"
)

type Module struct {
	Name   string
	Socket string
	DNA    *Common.P2PDNA
}

type LocalConnection struct {
	// spawn new listener, channel into output manager
	Socket string
	Target *Common.Peer
	Out    *OutboundDataChannel
}

func (c *LocalConnection) Listen() {

}

type OutboundData struct {
}

type OutboundDataChannel chan OutboundData

type Anevonet struct {
	Engine   *xorm.Engine
	Modules  map[string]*Module
	Services map[string]*irpc.Service
	// PeerConnection initiated through algorithms
	Connections  map[*Common.Peer]*LocalConnection
	OnlinePeers  []*Common.Peer
	Dir          string
	Outbound     OutboundDataChannel
	ID           Common.Peer
	Bootstrapped bool // true once we have atleast 1 connection
}

func (a *Anevonet) RegisterModule(module *irpc.Module) (*irpc.RegisterRes, error) {
	if _, ok := a.Modules[module.Name]; ok {
		return nil, errors.New("module already registered")
	}

	//TODO: fetch dna from db

	m := &Module{Name: module.Name, DNA: module.DNA, Socket: a.Dir + "/sockets/modules/" + module.Name}
	a.Modules[module.Name] = m

	res := &irpc.RegisterRes{DNA: m.DNA, Socket: m.Socket}
	return res, nil
}

func (a *Anevonet) RegisterService(s *irpc.Service) (bool, error) {
	if _, ok := a.Services[s.Name]; ok {
		return false, errors.New("service already registered")
	}
	a.Services[s.Name] = s
	return true, nil
}

func (a *Anevonet) RequestConnection(req *irpc.ConnectionReq) (*irpc.ConnectionRes, error) {
	log.Infof("RequestConnection\n")
	// check if we're already connected
	if _, ok := a.Connections[req.Target]; ok {
		return &irpc.ConnectionRes{Socket: a.Connections[req.Target].Socket}, nil
	}

	basepath := a.Dir + "/sockets/" + req.Module + "/connections/"
	err := os.MkdirAll(basepath, 0655)
	if err != nil {
		log.Fatal(err)
	}
	s := &LocalConnection{Target: req.Target, Socket: basepath + fmt.Sprintf("%s-%d", req.Target.IP, req.Target.Port)}
	a.Connections[req.Target] = s
	s.Out = &a.Outbound
	go s.Listen()

	res := &irpc.ConnectionRes{Socket: s.Socket}
	return res, nil
}

func (a *Anevonet) ShutdownConnection(req *irpc.ConnectionRes) error {
	return nil

}

func (a *Anevonet) Status() (*irpc.StatusRes, error) {
	log.Infof("Status\n")
	return &irpc.StatusRes{ID: &a.ID, OnlinePeers: int32(len(a.OnlinePeers))}, nil
}

func (a *Anevonet) BootstrapAlgorithm() (*irpc.BootstrapRes, error) {
	log.Infof("BootstrapAlogrithm\n")
	return &irpc.BootstrapRes{}, nil
}

func (a *Anevonet) BootstrapNetwork(peer *Common.Peer) (bool, error) {
	log.Infof("BootstrapNetwork (%d)\n", peer.Port)

	return false, nil
}

func pop(msg []string) (head, tail []string) {
	if msg[1] == "" {
		head = msg[:2]
		tail = msg[2:]
	} else {
		head = msg[:1]
		tail = msg[1:]
	}
	return
}

// Each worker task works on one request at a time and sends a random number
// of replies back, with random delays between replies:

func (a *Anevonet) InternalRPCWorker() {

	worker, _ := zmq3.NewSocket(zmq3.DEALER)
	defer worker.Close()
	worker.Connect("inproc://backend")

	for {
		//log.Info("waiting for data..")
		// The DEALER socket gives us the reply envelope and message
		msg, _ := worker.RecvMessage(0)
		/*msg, err := worker.RecvMessage(zmq3.DONTWAIT)
		if err != nil {
		continue
		}*/
		identity, content := pop(msg)
		//log.Infof("recv msg from %s: %s\n", identity, content)
		r := zmq.ThriftZMQChannel{}
		//c := make(chan []byte, 5)

		//r.ChanWriter = zmq.NewChanWriter(c)
		// drop first msg as it is framesize
		//log.Infof("new reader with n: %d\n", len(content))
		r.BufferReader = zmq.NewBufferReader(content)
		r.BufferWriter = zmq.NewBufferWriter()
		rpc.ServeRequest(thrift.NewServerCodec(zmq.NewFramedReadWriteCloser(r, 0), thrift.NewBinaryProtocol(true, false)))

		res := r.BufferWriter.Buf.Bytes()
		//log.Infof("sending back to (%s) (%d): %s\n", identity, len(res), res)
		worker.SendMessage(identity, res)
		//log.Info("done sending!!!")
	}
}

func (a *Anevonet) InternalRPC(port int) {

	rpc.RegisterName("Thrift", &irpc.InternalRpcServer{a})

	frontend := zmq.NewZMQConnection("Daemon", port, zmq.Server)

	// Backend socket talks to workers over inproc
	backend, _ := zmq3.NewSocket(zmq3.DEALER)
	defer backend.Close()
	backend.Bind("inproc://backend")

	// Launch pool of worker threads, precise number is not critical
	for i := 0; i < 2; i++ {
		go a.InternalRPCWorker()
	}

	// Connect backend to frontend via a proxy
	err := zmq3.Proxy(frontend.Sock, backend, nil)
	log.Fatalln("Proxy interrupted:", err)

}

var backendport int
var p2pport int
var dir string

func main() {
	flag.IntVar(&backendport, "port", 9000, "the port to start an instance of anevonet")
	flag.IntVar(&p2pport, "p2pport", 10000, "the port to run the p2p magic")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.Parse()

	log.Infof("staring daemon on %d and %d in %s\n", backendport, p2pport, dir)

	d, _ := filepath.Abs(dir)
	a := Anevonet{Dir: d, Modules: make(map[string]*Module), Connections: make(map[*Common.Peer]*LocalConnection), ID: Common.Peer{Port: int32(p2pport)}}

	engine, err := xorm.NewEngine("sqlite3", dir+"/anevonet.db")
	defer engine.Close()
	a.Engine = engine
	if err != nil {
		panic(err)
	}

	go a.InternalRPC(backendport)

	// read config file
	// open database

	// start evolver
	// start external listener
	// start connection manager
	// start internal broker

	// wait for SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for _ = range c {
		break
	}

	log.Info("stopping anevonet daemon\n")
	log.Flush()
}
