package libanevonet

import (
	"Common"
	rpc "anevonet_rpc"

	"flag" //flag "github.com/ogier/pflag"
	log "github.com/golang/glog"
	"github.com/samuel/go-thrift/thrift"
	zmq "libzmqthrift"
	//"net"
)

/* client side connection to the daemon */
type client struct {
	Zmq *zmq.ZmqConnection
}

/* Server, used for messages from the daemon*/
type server struct {
}

/* we abstract this and export an AnEvo Connection which is bidirectional */
type AnEvoConnection struct {
	Name        string
	Client      client
	Server      server
	Rpc         rpc.InternalRpcClient
	Connections map[*Common.Peer]zmq.RPCClient
}

func (a *AnEvoConnection) Connect(port int) {
	log.Infof("Connecting on: %d", port)
	a.Client.Zmq = zmq.NewZMQConnection(a.Name, port, zmq.Client)

	client := a.Client.Zmq.NewThriftClient()

	a.Rpc = rpc.InternalRpcClient{client}
}

func (*AnEvoConnection) ContinueRunning(dna Common.P2PDNA) bool {
	return true
}

func (a *AnEvoConnection) Bootstrap() ([]*Common.Peer, error) {

	r, err := a.Rpc.BootstrapAlgorithm()
	if err != nil {
		return nil, err
	}
	return r.Peers, nil
}

func (a *AnEvoConnection) GetPeerConnection(p *Common.Peer) zmq.RPCClient {
	// are we already connected to that peer?
	if val, ok := a.Connections[p]; ok {
		return val
	}

	// make a new connection
	r, err := a.Rpc.RequestConnection(&rpc.ConnectionReq{Target: p, Module: a.Name})
	if err != nil {
		panic(err)
	}

	/*conn, err := net.Dial("unix", r.Socket)
	if err != nil {
		panic(err)
	}

	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false), false)
	*/

	z := zmq.NewZMQUnixConnection(r.Socket)
	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(z, 0), thrift.NewBinaryProtocol(false, false), false)

	return client
}

func (a *AnEvoConnection) Register(rootdna Common.P2PDNA, dna *Common.P2PDNA) string {
	log.Infof("Register Module: %s", a.Name)
	r, err := a.Rpc.RegisterModule(&rpc.Module{Name: a.Name, DNA: &rootdna})
	if err != nil {
		panic(err)
	}
	// check db for better dna
	dna = r.DNA
	return r.Socket
}

func NewModule(name string) *AnEvoConnection {
	c := &AnEvoConnection{}
	var port int
	flag.IntVar(&port, "port", 9000, "port of the daemon")
	flag.Parse()
	c.Name = name
	c.Connect(port)

	return c
}
