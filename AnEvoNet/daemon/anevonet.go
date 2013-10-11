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


6. NAT traversal
6.1 NAT-pmp http://code.google.com/p/go-nat-pmp/
6.2 UpnP https://github.com/jackpal/Taipei-Torrent/blob/master/upnp.go

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
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	erpc "p2p_rpc"
	"path/filepath"
	"strings"
)

type Module struct {
	Name   string
	Socket string
	DNA    *Common.P2PDNA
}

type LocalConnection struct {
	// spawn new listener, channel into output manager
	Socket string
	Module string
	Target *Common.Peer
	Out    OutboundDataChannel
}

func (c *LocalConnection) Listen() error {
	server, _ := zmq3.NewSocket(zmq3.REP)
	defer server.Close()
	fp := zmq.FixUnixSocketPath(c.Socket)
	err := server.Bind(fmt.Sprintf("ipc://%s", fp))

	if err != nil {
		log.Error("cannot listen on socket", err)
		return err
	}
	log.Infof("Listening on %s\n", fp)
	go func() {
		for {
			//log.Info("waiting for data..")
			// The DEALER socket gives us the reply envelope and message
			msg, _ := server.RecvMessage(0)
			/*msg, err := worker.RecvMessage(zmq3.DONTWAIT)
			if err != nil {
			continue
			}*/
			identity, content := pop(msg)

			log.Infof("recv msg from %s: %s\n", identity, content)
			out := OutboundData{Target: c.Target, Module: c.Module, Data: []byte(strings.Join(content, ""))}
			c.Out <- out
			//log.Infof("sending back to (%s) (%d): %s\n", identity, len(res), res)
			server.SendMessage(identity, "ACK")
			//log.Info("done sending!!!")
		}
	}()
	return nil
}

type RemoteConnection struct {
	Peer       *Common.Peer
	Us         *Anevonet
	Connection *net.Conn
	Out        OutboundDataChannel
}

func (c *RemoteConnection) Connect() error {
	log.Infof("Remote Connect\n")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Peer.IP, c.Peer.Port))
	if err != nil {
		log.Error(err)
		return errors.New("cannot connect to peer")
	}

	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false), false)
	p2p := erpc.RemoteRpcClient{client}

	hi := erpc.HelloSYN{NodeID: c.Us.ID.ID, Version: Common.VERSION}
	res, err := p2p.Hello(&hi)
	if err != nil {
		log.Error(err)
		return errors.New("cannot rpc with peer")
	}
	//hisynack := erpc.HelloSYNACK{NodeID: c.Us.ID.ID, Transport}
	c.Connection = &conn
	fmt.Printf("Response: %+v\n", res)
	return c.Us.AddRemotePeer(c.Peer, c)
}

type OutboundData struct {
	Module string
	Data   []byte
	Target *Common.Peer
}

type OutboundDataChannel chan OutboundData

type PeerRConnection struct {
	Peer       *Common.Peer
	Connection *RemoteConnection
}
type Anevonet struct {
	Engine           *xorm.Engine
	Modules          map[string]*Module
	Services         map[string]*irpc.Service
	Connections      map[string]*PeerRConnection // peerid as key
	_rnd_Connections map[int]string
	Transports       map[Common.Transport]int32
	// Tunnels data from module specific socket to the remote peer
	Tunnels           map[*Common.Peer]*LocalConnection
	OnlinePeers       map[*Common.Peer]bool
	Dir               string
	Outbound          OutboundDataChannel
	ID                Common.Peer
	Bootstrapped      bool // true once we have atleast 1 connection
	InternalRPCServer *rpc.Server
	ExternalRPCServer *rpc.Server
}

func (a *Anevonet) Proxy() {
	log.Infof("proxy \n")
	//var d OutboundData
	for {
		select {
		case d := <-a.Outbound:
			log.Infof("Sending:", d)

		}
	}

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

	// connect to peer
	ok, _ := a.BootstrapNetwork(req.Target)
	if !ok {
		return nil, errors.New("Cannot Connect to peer")
	}

	// TODO: what if multiple modules access the same peer?
	// check if we're already connected
	if _, ok := a.Tunnels[req.Target]; ok {
		return &irpc.ConnectionRes{Socket: a.Tunnels[req.Target].Socket}, nil
	}

	basepath := a.Dir + "/sockets/" + req.Module + "/connections/"
	err := os.MkdirAll(basepath, 0766)
	if err != nil {
		log.Fatal(err)
	}
	s := &LocalConnection{Module: req.Module, Target: req.Target, Socket: basepath + fmt.Sprintf("%s-%d", req.Target.IP, req.Target.Port),
		Out: a.Connections[req.Target.ID].Connection.Out}
	a.Tunnels[req.Target] = s
	s.Out = a.Outbound
	// Listen locally for modules sending data
	err = s.Listen()
	if err != nil {
		return nil, errors.New("Cannot listen on local socket")
	}

	res := &irpc.ConnectionRes{Socket: s.Socket}
	return res, nil
}

func (a *Anevonet) ShutdownConnection(req *irpc.ConnectionRes) error {
	return nil

}

func (a *Anevonet) AddRemotePeer(p *Common.Peer, c *RemoteConnection) error {
	if _, ok := a.Connections[p.ID]; ok {
		return errors.New("Peer already registered")
	}
	a.Connections[p.ID] = &PeerRConnection{Peer: p, Connection: c}
	a._rnd_Connections[len(a._rnd_Connections)] = p.ID
	log.Infof("Added %s-%s:%d --- %d\n", p.ID, p.IP, p.Port, len(a.Connections))
	return nil
}

func (a *Anevonet) ConnectedToPeer(p *Common.Peer) bool {
	_, ok := a.Connections[p.ID]
	return ok
}

func (a *Anevonet) ConnectRemotePeer(p *Common.Peer) error {
	c := &RemoteConnection{Peer: p, Us: a}
	return c.Connect()

}

func (a *Anevonet) Status() (*irpc.StatusRes, error) {
	log.Infof("Status\n")
	return &irpc.StatusRes{ID: &a.ID, OnlinePeers: int32(len(a.OnlinePeers))}, nil
}

func (a *Anevonet) BootstrapAlgorithm() (*irpc.BootstrapRes, error) {
	log.Infof("BootstrapAlgorithm\n")

	if len(a.Connections) < 1 {
		return &irpc.BootstrapRes{}, nil
	} else {
		//pick random connection
		nC := rand.Intn(len(a._rnd_Connections))
		id := a._rnd_Connections[nC]
		var peers = []*Common.Peer{a.Connections[id].Peer}
		return &irpc.BootstrapRes{Peers: peers}, nil
	}
}

func (a *Anevonet) BootstrapNetwork(peer *Common.Peer) (bool, error) {
	log.Infof("BootstrapNetwork (%d)\n", peer.Port)
	if a.ConnectedToPeer(peer) {
		log.Infof("We're already conntected to (%d)\n", peer.Port)
		return true, nil
	}
	err := a.ConnectRemotePeer(peer)
	if err == nil {
		return true, err
	} else {
		return false, err
	}
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
		a.InternalRPCServer.ServeRequest(thrift.NewServerCodec(zmq.NewFramedReadWriteCloser(r, 0), thrift.NewBinaryProtocol(true, false)))

		res := r.BufferWriter.Buf.Bytes()
		//log.Infof("sending back to (%s) (%d): %s\n", identity, len(res), res)
		worker.SendMessage(identity, res)
		//log.Info("done sending!!!")
	}
}

func (a *Anevonet) InternalRPC(port int) {
	a.InternalRPCServer = rpc.NewServer()
	a.InternalRPCServer.RegisterName("Thrift", &irpc.InternalRpcServer{a})

	frontend := zmq.NewZMQConnection("Daemon", port, zmq.Server)

	// Backend socket talks to workers over inproc
	backend, _ := zmq3.NewSocket(zmq3.DEALER)
	defer backend.Close()
	backend.Bind("inproc://backend")

	// Launch pool of worker threads, precise number is not critical
	for i := 0; i < 2; i++ {
		go a.InternalRPCWorker()
	}
	var err error
	// Connect backend to frontend via a proxy
	for err != errors.New("interrupted system call") {
		err = zmq3.Proxy(frontend.Sock, backend, nil)
		log.Errorln("Proxy interrupted:", err)
	}
	log.Fatalln("Proxy errord:", err)
}

type P2PRPCServer struct {
	ae *Anevonet
}

func (s *P2PRPCServer) Hello(hello *erpc.HelloSYN) (*erpc.HelloSYNACK, error) {
	log.Infof("Hello from %s\n", hello.Version)
	resp := erpc.HelloSYNACK{NodeID: s.ae.ID.ID, Version: Common.VERSION, SupportedTransport: s.ae.Transports}
	return &resp, nil
}

func (s *P2PRPCServer) Connect(c *erpc.ConnectSYN) (bool, error) {
	log.Infof("connect from %s\n", c.NodeID)
	//resp := erpc.HelloSYNACK{NodeID: s.ae.ID.ID, Version: Common.VERSION}
	return true, nil
}

func (a *Anevonet) P2PRPC(port int32) {

	a.TransportTCP(port)

}

func (a *Anevonet) TransportTCP(port int32) {
	a.Transports[Common.TransportTCP] = port

	ps := &P2PRPCServer{ae: a}

	a.ExternalRPCServer = rpc.NewServer()
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		panic(err)
	}

	a.ExternalRPCServer.RegisterName("Thrift", &erpc.RemoteRpcServer{ps})
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("ERROR: %+v\n", err)
			continue
		}
		fmt.Printf("New P2P connection %+v\n", conn)
		go a.ExternalRPCServer.ServeCodec(thrift.NewServerCodec(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false)))
	}

}

var backendport int
var p2pport int
var dir string

func main() {
	flag.IntVar(&backendport, "rpc-port", 9000, "port of the local rpc service")
	flag.IntVar(&p2pport, "p2p-port", 10000, "port of the p2p service")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.Parse()

	log.Infof("staring daemon on %d and %d in %s\n", backendport, p2pport, dir)

	d, _ := filepath.Abs(dir)
	a := Anevonet{Dir: d, Modules: make(map[string]*Module),
		Tunnels:          make(map[*Common.Peer]*LocalConnection),
		Connections:      make(map[string]*PeerRConnection),
		_rnd_Connections: make(map[int]string),
		Transports:       make(map[Common.Transport]int32),
		ID:               Common.Peer{Port: int32(p2pport), ID: "JAJAJAJAJAJ"},
		Outbound:         make(OutboundDataChannel)}

	engine, err := xorm.NewEngine("sqlite3", dir+"/anevonet.db")
	defer engine.Close()
	a.Engine = engine
	if err != nil {
		panic(err)
	}

	go a.InternalRPC(backendport)
	go a.P2PRPC(int32(p2pport))
	go a.Proxy()
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
