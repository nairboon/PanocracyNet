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
	"github.com/samuel/go-thrift/thrift"
	"net"
	"net/rpc"
	"path/filepath"
)

type Module struct {
	Name   string
	Socket string
	DNA    Common.DNA
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
	Engine      *xorm.Engine
	Modules     map[string]*Module
	Connections map[*Common.Peer]*LocalConnection
	Dir         string
	Outbound    OutboundDataChannel
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

func (a *Anevonet) RequestConnection(req *irpc.ConnectionReq) (*irpc.ConnectionRes, error) {
	// check if we're already connected
	if _, ok := a.Connections[req.Target]; ok {
		return &irpc.ConnectionRes{Socket: a.Connections[req.Target].Socket}, nil
	}

	//TODO: fetch dna from db
	s := &LocalConnection{Target: req.Target, Socket: a.Dir + "/sockets/connections/" + fmt.Sprintf("%s-%d", req.Target.IP, req.Target.Port)}
	a.Connections[req.Target] = s
	s.Out = &a.Outbound
	go s.Listen()

	res := &irpc.ConnectionRes{Socket: s.Socket}
	return res, nil
}

func (a *Anevonet) InternalRPC(port int) {

	rpc.RegisterName("Thrift", &irpc.InternalRpcServer{a})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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

var port int
var dir string

func main() {
	flag.IntVar(&port, "port", 9000, "the port to start an instance of anevonet")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.Parse()

	log.Info("staring daemon on %d in %s\n", port, dir)

	d, _ := filepath.Abs(dir)
	a := Anevonet{Dir: d}
	engine, err := xorm.NewEngine("sqlite3", dir+"/anevonet.db")
	defer engine.Close()
	a.Engine = engine
	if err != nil {
		panic(err)
	}

	go a.InternalRPC(port)
	// declare channels
	//api_calls := make(APICallChannel, 5)

	// read config file
	// open database

	// start evolver
	// start external listener
	// start connection manager
	// start internal broker

	log.Info("stopping anevonet daemon\n")
	log.Flush()
}
