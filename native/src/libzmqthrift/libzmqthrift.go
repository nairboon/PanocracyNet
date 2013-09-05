package libzmqthrift

import (
	"fmt"
	"github.com/samuel/go-thrift/thrift"
	zmq "github.com/vaughan0/go-zmq"
	"io"
)

type RPCClient interface {
	Call(method string, request interface{}, response interface{}) error
}

const (
	Client zmq.SocketType = zmq.Req
	Server                = zmq.Rep
)

/* zeromq connection */
type ZmqConnection struct {
	Ctx   *zmq.Context
	Chans *zmq.Channels
	Sock  *zmq.Socket

	*chanReader
	*chanWriter
}

func (z *ZmqConnection) Close() error {
	z.Ctx.Close()
	z.Sock.Close()
	z.Chans.Close()
	return nil
}

func (z *ZmqConnection) NewThriftClient() RPCClient {
	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(z, 0), thrift.NewBinaryProtocol(true, false), false)

	return client
}

func NewZMQConnection(port int, t zmq.SocketType) *ZmqConnection {
	c := &ZmqConnection{}
	var err error
	c.Ctx, err = zmq.NewContext()
	if err != nil {
		panic(err)
	}

	c.Sock, err = c.Ctx.Socket(t)
	if err != nil {
		panic(err)
	}

	if err = c.Sock.Bind(fmt.Sprintf("tcp://*:%d", port)); err != nil {
		panic(err)
	}
	c.Chans = c.Sock.Channels()
	c.chanReader = newChanReader(c.Chans.In())

	return c
}

// chanReader receives on the channel when its
// Read method is called. Extra data received is
// buffered until read.
type chanReader struct {
	resbuf []byte
	buf    [][]byte
	c      <-chan [][]byte
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
	r.buf = r.buf[:0]
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
