package libanevonet

import (
	"Common"
	rpc "anevonet_rpc"
	flag "github.com/ogier/pflag"
	"github.com/samuel/go-thrift/thrift"
	zmq "libzmqthrift"
	"log"
	"net"
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
	Client      client
	Server      server
	Rpc         rpc.InternalRpcClient
	Connections map[*Common.Peer]zmq.RPCClient
}

func (a *AnEvoConnection) Connect(port int) {
	log.Printf("Connecting on: %d", port)
	a.Client.Zmq = zmq.NewZMQConnection(port, zmq.Client)

	client := a.Client.Zmq.NewThriftClient()

	a.Rpc = rpc.InternalRpcClient{client}
}

func (*AnEvoConnection) ContinueRunning(dna Common.DNA) bool {
	return true
}

func (a *AnEvoConnection) GetPeerConnection(p *Common.Peer) zmq.RPCClient {
	// are we already connected to that peer?
	if val, ok := a.Connections[p]; ok {
		return val
	}

	// make a new connection
	r, err := a.Rpc.RequestConnection(&rpc.ConnectionReq{Target: p})
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("unix", r.Socket)
	if err != nil {
		panic(err)
	}

	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false), false)

	return client
}

func (*AnEvoConnection) Register(name string, rootdna Common.DNA, dna Common.DNA) string {

	// check db for better dna
	dna = rootdna
	return "SOCKETFILE"
}

func NewConnection() *AnEvoConnection {
	c := &AnEvoConnection{}
	var port int
	flag.IntVar(&port, "port", 9000, "port of the daemon")
	flag.Parse()

	c.Connect(port)

	return c
}
