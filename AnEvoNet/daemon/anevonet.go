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
	zmq3 "github.com/pebbe/zmq4"
	"github.com/samuel/go-thrift/thrift"
	zmq "libzmqthrift"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	erpc "p2p_rpc"
	"path/filepath"
	//"strings"
	"runtime"
	"time"
)

type Module struct {
	Name   string
	Socket string
	DNA    *Common.P2PDNA
}

func (m *Module) Listen() error {

	//server, _ := zmq3.NewSocket(zmq3.REP)

	//fp := zmq.FixUnixSocketPath(m.Socket)
	//err := server.Bind(fmt.Sprintf("ipc://%s", fp))

	//if err != nil {
	//	log.Errorf("cannot listen on socket: %s", err)
	//	return err
	//}

	return nil
}

/*
 send message over the module socket, this should trigger an RPC call
*/
func (m *Module) SendMessage(msg string, rch chan []byte, ech chan error) (string, error) {
	log.Infof("send message to local module")
	//server, _ := zmq3.NewSocket(zmq3.REP)

	//fp := zmq.FixUnixSocketPath(m.Socket)
	//err := server.Bind(fmt.Sprintf("ipc://%s", fp))

	//if err != nil {
	//	log.Errorf("cannot listen on socket: %s", err)
	//	return err
	//}

	//con := zmq.NewZMQUnixConnection(m.Socket)

	con, err := net.Dial("tcp", m.Socket)
	if err != nil {
		panic(err)
	}

	//con := zmq.NewFramedReadWriteCloser(z, 0)
	con.Write([]byte(msg))
	log.Infof("Sent message over ZMQ to %s", m.Socket)

	//con.Flush()
	//con.Send()
	//z.Send()

	res := make([]byte, 200)
	log.Infof("Going to read on SOCKET")

	/*ch := make(chan []byte)
	eCh := make(chan error)

	// Start a goroutine to read from our net connection
	go func(ch chan []byte, eCh chan error) {
		for {
			// try to read the data
			data := make([]byte, 512)
			_, err := con.Read(data)
			if err != nil {
				// send an error if it's encountered
				eCh <- err
				return
			}
			// send data if we read some.
			ch <- data
		}
	}(ch, eCh)

	ok := true
	for ok {
		select {
		// This case means we recieved data on the connection
		case data := <-ch:
			log.Infof("DATA %s", data)

			// Do something with the data
			res = data
		// This case means we got an error and the goroutine has finished
		case err := <-eCh:
			log.Infof("ERROR %v", err)

			// handle our error then exit for loop
			break
		// This will timeout on the read.
		case _ = <-time.Tick(time.Second):
			log.Infof("Second TIMEOUT")
			ok = false
			// do nothing? this is just so we can time out if we need to.
			// you probably don't even need to have this here unless you want
			// do something specifically on the timeout.
			break
		}
	}*/
	con.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := con.Read(res)
	log.Infof("Got answer l(%d) %#v   %#v", n, err, res)

	mp := runtime.GOMAXPROCS(0)

	log.Infof("some stats %d -- %d", mp, runtime.NumGoroutine())
	if n == 0 {
		panic(0)
		ech <- errors.New("Connection timedout")
		rch <- []byte("")
	}
	res = res[:n]
	//con.Close()
	rch <- res
	ech <- nil
	log.Infof("sent res to channels")

	time.Sleep(2 * time.Second)

	return string(res[:n]), nil
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

	fp := zmq.FixUnixSocketPath(c.Socket)
	err := server.Bind(fmt.Sprintf("ipc://%s", fp))

	if err != nil {
		log.Errorf("cannot listen on socket: %s", err)
		return err
	}
	log.Infof("Listening on %s\n", fp)
	go func() {
		for {
			//log.Info("waiting for data..")
			// The DEALER socket gives us the reply envelope and message
			//msg, _ := server.RecvMessage(0)
			//msg, err := server.RecvMessage(0)
			msg, err := server.Recv(0)
			//fmt.Println("Received ", msg)
			if err != nil {
				//log.Error(err)
				//break
				continue
			}

			identity, _ := server.GetIdentity()

			//content := msg[0]
			//identity, content := pop(msg)

			log.Infof("recv msg from %v -> %s\n", identity, msg)
			//out := OutboundData{Target: c.Target, Module: c.Module, Data: []byte(strings.Join(content, ""))}
			out := OutboundData{Target: c.Target, Module: c.Module, Data: []byte(msg)}

			out.Answer = make(chan []byte)
			//TODO: timeout
			c.Out <- out
			log.Infof("sent to outbound, wait for result...")

			//answer := make([]byte, 1024)

			answer := <-out.Answer

			log.Infof("we got an answer: %v", answer)

			//log.Infof("sending back to (%s) (%d): %s\n", identity, len(res), res)
			server.Send(string(answer), 0)
			//server.Send("", 0)
			log.Info("done sending!!!")
		}
		defer server.Close()
	}()
	return nil
}

type RemoteConnection struct {
	Peer       *Common.Peer
	Us         *Anevonet
	Connection *net.Conn
	Out        OutboundDataChannel
	Client     erpc.RemoteRpcClient
}

func (c *RemoteConnection) Send(data OutboundData) (*erpc.Message, error) {

	req := erpc.Message{Module: data.Module, Payload: string(data.Data)}
	res, err := c.Client.SendMessage(&req)
	if err != nil {
		log.Error(err)
		return &erpc.Message{}, errors.New("cannot rpc with peer")
	}

	return res, nil
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
	c.Client = p2p
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
	Answer chan []byte
}

type OutboundDataChannel chan OutboundData
type QuitChannel chan bool
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
			log.Infof("Going to send: %v", d)

			con, ok := a.Connections[d.Target.ID]
			if !ok {
				log.Errorf("we have no connection to %v", d.Target)
			}
			res, err := con.Connection.Send(d)
			if err != nil {
				log.Errorf("Failed: %v", err)
			}

			log.Infof("got response: %v, forward..", res)
			d.Answer <- []byte(res.Payload)
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

	m.Socket = fmt.Sprintf(":%d", a.ID.Port+23)
	basepath := a.Dir + "/sockets/modules/"
	err := os.MkdirAll(basepath, 0766)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Listen()
	if err != nil {
		log.Fatal(err)
	}

	res := &irpc.RegisterRes{DNA: m.DNA, Socket: m.Socket, ID: &a.ID}
	log.Infof("RegisterModule (%v)", res)
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
	s := &LocalConnection{Module: req.Module, Target: req.Target, Socket: basepath + fmt.Sprintf("%s-%d", req.Target.IP, req.Target.Port)}
	/*Out: a.Connections[req.Target.ID].Connection.Out*/
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
	log.Infof("BootstrapAlgorithm %v\n", a.ID)

	if len(a.Connections) < 1 {
		log.Infof("BootstrapAlgorithm: we have %d connections...\n", len(a.Connections))
		return &irpc.BootstrapRes{}, nil
	} else {
		//pick random connection

		nC := rand.Intn(len(a._rnd_Connections))
		id := a._rnd_Connections[nC]
		log.Infof("BootstrapAlgorithm: pick random connection %v\n", id)
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

func (a *Anevonet) InternalRPCWorker(quit QuitChannel) {

	worker, _ := zmq3.NewSocket(zmq3.DEALER)
	defer worker.Close()
	worker.Connect("inproc://backend")

	for {
		select {
		case <-quit:
			return
		default:

		}
		//time.Sleep(10*time.Millisecond)
		//log.Info("waiting for data..")
		// The DEALER socket gives us the reply envelope and message
		msg, err := worker.RecvMessage(0)
		/*msg, err := worker.RecvMessage(zmq3.DONTWAIT)*/
		if err != nil {
			continue
		}
		identity, content := pop(msg)
		log.Infof("recv msg from %s: %s\n", identity, content)
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

func (a *Anevonet) InternalRPC(port int, quit QuitChannel) {
	a.InternalRPCServer = rpc.NewServer()
	a.InternalRPCServer.RegisterName("Thrift", &irpc.InternalRpcServer{a})

	frontend, err := zmq.NewZMQConnection("Daemon", port, zmq.Server)
	defer frontend.Close()
	if err != nil {
		log.Fatalln("Cannot bind frontend:", err)
	}
	// Backend socket talks to workers over inproc
	backend, _ := zmq3.NewSocket(zmq3.DEALER)
	defer backend.Close()
	err = backend.Bind("inproc://backend")
	if err != nil {
		log.Fatalln("Cannot bind backend:", err)
	}
	// Launch pool of worker threads, precise number is not critical
	for i := 0; i < 2; i++ {
		go a.InternalRPCWorker(quit)
	}

	// Connect backend to frontend via a proxy
	log.Infof("Start Proxy\n")
	for err != errors.New("interrupted system call") {
		err = zmq3.Proxy(frontend.Sock, backend, nil)
		log.Errorln("Proxy interrupted:", err)
		select {
		case <-quit:
			for i := 0; i < 2; i++ {
				quit <- true
			}
			return
		default:

		}
		time.Sleep(500 * time.Millisecond)

	}
	log.Fatalln("Proxy errord:", err)
}

type P2PRPCServer struct {
	ae *Anevonet
}

func (s *P2PRPCServer) Hello(hello *erpc.HelloSYN) (*erpc.HelloSYNACK, error) {
	log.Infof("RPC:Hello from %s\n", hello.Version)
	resp := erpc.HelloSYNACK{NodeID: s.ae.ID.ID, Version: Common.VERSION, SupportedTransport: s.ae.Transports}
	return &resp, nil
}

func (s *P2PRPCServer) Connect(c *erpc.ConnectSYN) (bool, error) {
	log.Infof("RPC:connect from %s\n", c.NodeID)
	//resp := erpc.HelloSYNACK{NodeID: s.ae.ID.ID, Version: Common.VERSION}
	return true, nil
}

func (s *P2PRPCServer) SendMessage(m *erpc.Message) (*erpc.Message, error) {
	log.Infof("RPC:send message %v\n", m)
	//resp := erpc.HelloSYNACK{NodeID: s.ae.ID.ID, Version: Common.VERSION}

	// lookup module
	if _, ok := s.ae.Modules[m.Module]; !ok {
		log.Errorf("module %v does not exist in %v", m.Module, s.ae.Modules)
		return &erpc.Message{}, errors.New("Module does not exist")
	}
	module := s.ae.Modules[m.Module]

	rch := make(chan []byte)
	eCh := make(chan error)
	go module.SendMessage(m.Payload, rch, eCh)
	log.Infof("checking channels ...")

	var res []byte
	var err error
	stop := false
	for !stop {
		select {
		// This case means we recieved data on the connection
		case res = <-rch:
			log.Infof("A res: %v", res)
			stop = true
			break

		case err = <-eCh:
			log.Infof("An err: %v", err)
			stop = true
			break
		}
	}

	log.Infof("We received msg from module %#v %#v", res, err)

	return &erpc.Message{Module: m.Module, Payload: string(res)}, err
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
var id string

func main() {
	flag.IntVar(&backendport, "rpc-port", 9000, "port of the local rpc service")
	flag.IntVar(&p2pport, "p2p-port", 10000, "port of the p2p service")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.StringVar(&id, "id", "0000", "peerid")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Infof("staring daemon on %d and %d in %s\n", backendport, p2pport, dir)

	d, _ := filepath.Abs(dir)
	a := Anevonet{Dir: d, Modules: make(map[string]*Module),
		Tunnels:          make(map[*Common.Peer]*LocalConnection),
		Connections:      make(map[string]*PeerRConnection),
		_rnd_Connections: make(map[int]string),
		Transports:       make(map[Common.Transport]int32),
		ID:               Common.Peer{Port: int32(p2pport), ID: id},
		Outbound:         make(OutboundDataChannel)}

	engine, err := xorm.NewEngine("sqlite3", dir+"/anevonet.db")
	defer engine.Close()
	a.Engine = engine
	if err != nil {
		panic(err)
	}
	quit := make(chan bool)
	go a.InternalRPC(backendport, quit)
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
	quit <- true
	log.Info("stopping anevonet daemon\n")
	defer zmq3.Term()

	log.Flush()
}
