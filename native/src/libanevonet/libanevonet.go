package libanevonet

import (
	flag "github.com/ogier/pflag"
	"log"
	"Common"
	"fmt"
 rpc "anevonet_rpc"
"io"
	zmq "github.com/vaughan0/go-zmq"
	"github.com/samuel/go-thrift/thrift"
	"net"
)

/* zeromq connection */
type zmqConnection struct {
 Ctx *zmq.Context
 Chans *zmq.Channels
 Sock *zmq.Socket

   *chanReader
   *chanWriter
}

func newZMQConnection(port int) *zmqConnection {
  c := &zmqConnection{}
var err error
c.Ctx, err = zmq.NewContext()
if err != nil {
  panic(err)
}
defer c.Ctx.Close()

c.Sock, err = c.Ctx.Socket(zmq.Rep)
if err != nil {
  panic(err)
}
defer c.Sock.Close()



if err = c.Sock.Bind(fmt.Sprintf("tcp://*:%d",port)); err != nil {
  panic(err)
}
c.Chans = c.Sock.Channels()
defer c.Chans.Close()

c.chanReader = newChanReader(c.Chans.In())

 return c
}

/* client side connection to the daemon */
type client struct {
 zmq *zmqConnection
}


/* Server, used for messages from the daemon*/
type server struct {

}

type RPCClient interface {
	Call(method string, request interface{}, response interface{}) error
}

/* we abstract this and export an AnEvo Connection which is bidirectional */
type AnEvoConnection struct {
 Client client
 Server server
 Rpc rpc.LocalRpcClient
 Connections map[Common.Peer]RPCClient
}

func (a *AnEvoConnection) Connect(port int) {
 log.Printf("Connecting on: %d",port)
  a.Client.zmq = newZMQConnection(port)

client := thrift.NewClient(thrift.NewFramedReadWriteCloser(a.Client.zmq, 0), thrift.NewBinaryProtocol(true, false),false)

 a.Rpc = rpc.LocalRpcClient{client}
}

func (*AnEvoConnection) ContinueRunning(dna Common.DNA) bool{
 return true
}


func (a *AnEvoConnection) GetPeerConnection(p Common.Peer) RPCClient {
// are we already connected to that peer?
if val,ok := a.Connections[p]; ok {
    return val
}

// make a new connection
 r,err  := a.Rpc.RequestConnection(&rpc.ConnectionReq{Target: &p})
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

func (*AnEvoConnection) Register(name string, rootdna Common.DNA, dna Common.DNA){

// check db for better dna
dna = rootdna
}

func NewConnection() *AnEvoConnection {
  c := &AnEvoConnection{}
  var port int
 	flag.IntVar(&port, "port", 9000, "port of the daemon")
	flag.Parse()

 c.Connect(port)

return c
}

// chanReader receives on the channel when its
// Read method is called. Extra data received is
// buffered until read.
type chanReader struct {
        resbuf []byte
        buf [][]byte
        c   <-chan [][]byte
}

func newChanReader(c <-chan [][]byte) *chanReader {
        return &chanReader{c: c}
}

func (r *chanReader) Read(buf []byte) (int, error) {
        for len(r.resbuf) == 0 {
                var ok bool
                r.buf, ok = <-r.c
                if !ok {
                        return 0, io.EOF
                }
		for _, part := range r.buf[:len(r.buf)] {
			r.resbuf = append(r.resbuf, part...)
       		 }

		
	}
	r.buf=r.buf[:0]
        n := copy(buf, r.resbuf)
	r.resbuf = r.resbuf[n:]
        return n, nil
}

func (r *chanReader) Close() error {
	return nil
}

// chanWriter writes on the channel when its
// Write method is called.
type chanWriter struct {
        c chan<- [][]byte
}

func newChanWriter(c chan<- [][]byte) *chanWriter {
        return &chanWriter{c: c}
}
func (w *chanWriter) Write(buf []byte) (n int, err error) {
        b := make([][]byte, 1)
	b[0] = make([]byte, len(buf))
        copy(b[0], buf)
        w.c <- b
        return len(buf), nil
}
 
