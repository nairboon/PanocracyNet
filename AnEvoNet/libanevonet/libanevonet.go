package libanevonet

import (
	"Common"
	rpc "anevonet_rpc"
	//"errors"
	"flag" //flag "github.com/ogier/pflag"
	log "github.com/golang/glog"
	golog "log"
	//"github.com/samuel/go-thrift/thrift"
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

func (a *AnEvoConnection) Connect(port int) error {
	log.Infof("Connecting on: %d", port)
	var err error
	a.Client.Zmq, err = zmq.NewZMQConnection(a.Name, port, zmq.Client)
	if err != nil {
		return err
	}
	client := a.Client.Zmq.NewThriftClient()

	a.Rpc = rpc.InternalRpcClient{client}
	return nil
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

func (a *AnEvoConnection) GetPeerConnection(p *Common.Peer) (zmq.RPCClient, error) {
	// are we already connected to that peer?
	if val, ok := a.Connections[p]; ok {
		return val, nil
	}

	// make a new connection
	r, err := a.Rpc.RequestConnection(&rpc.ConnectionReq{Target: p, Module: a.Name})
	if err != nil {
		return nil, err
	}

	/*conn, err := net.Dial("unix", r.Socket)
	if err != nil {
		panic(err)
	}

	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(conn, 0), thrift.NewBinaryProtocol(true, false), false)
	*/
	log.Infof("Connecting to socket %s", r.Socket)
	z := zmq.NewZMQUnixConnection(r.Socket)
	//client := thrift.NewClient(thrift.NewFramedReadWriteCloser(z, 0), thrift.NewBinaryProtocol(false, false), false)
	client := z.NewThriftClient()
	return client, nil
}

func (a *AnEvoConnection) Register(rootdna Common.P2PDNA) (*rpc.RegisterRes, Common.P2PDNA, error) {
	log.Infof("Register Module: %s", a.Name)
	r, err := a.Rpc.RegisterModule(&rpc.Module{Name: a.Name, DNA: &rootdna})
	log.Infof("done registering")
	if err != nil {
		return &rpc.RegisterRes{}, Common.P2PDNA{}, err
	}
	// check db for better dna
	dna := *r.DNA

	golog.Printf("DNA is %#v vs %#v", dna, rootdna)

	return r, dna, nil
}

var flagSet = false
var port int

func NewModule(name string) (*AnEvoConnection, error) {
	c := &AnEvoConnection{}
	if !flagSet {

		flag.IntVar(&port, "port", 9000, "port of the daemon")
		flag.Parse()
		flagSet = true
	}
	c.Name = name
	err := c.Connect(port)

	return c, err
}
