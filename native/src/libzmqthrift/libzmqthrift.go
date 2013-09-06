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
	Client zmq.SocketType = zmq.Dealer
	Server                = zmq.Router
)

/* zeromq connection */
type ZmqConnection struct {
	Ctx   *zmq.Context
	Chans *zmq.Channels
	Sock  *zmq.Socket

	*chanReader
	*ChanWriter
}

type ThriftZMQChannel struct {
	*BufferReader
	*ChanWriter
}

func (t ThriftZMQChannel) Close() error {
	//t.R.Close()
	return nil
} /*
func (z *ZmqConnection) Close() error {
	z.Ctx.Close()
	z.Sock.Close()
	z.Chans.Close()
	return nil
}*/

func (z *ZmqConnection) NewThriftClient() RPCClient {
	client := thrift.NewClient(thrift.NewFramedReadWriteCloser(z, 0), thrift.NewBinaryProtocol(false, false), false)
	//client := thrift.NewClient(z, thrift.NewBinaryProtocol(false, false), false)
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
	if t == Server {
		err = c.Sock.Bind(fmt.Sprintf("tcp://*:%d", port))
	} else {
		err = c.Sock.Connect(fmt.Sprintf("tcp://localhost:%d", port))
	}

	if err != nil {
		panic(err)
	}
	//if t == Client {
	c.Chans = c.Sock.ChannelsBuffer(5)
	c.chanReader = NewChanReader(c.Chans.In())
	c.ChanWriter = NewChanWriter(c.Chans.Out())

	//}

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

func NewChanReader(c <-chan [][]byte) *chanReader {
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

// ChanWriter writes on the channel when its
// Write method is called.
type ChanWriter struct {
	c chan<- [][]byte
}

func NewChanWriter(c chan<- [][]byte) *ChanWriter {
	return &ChanWriter{c: c}
}
func (w *ChanWriter) Write(buf []byte) (n int, err error) {
	b := make([][]byte, 1)
	b[0] = make([]byte, len(buf))
	copy(b[0], buf)
	w.c <- b
	fmt.Printf("wrote to chan: %s", b)
	//panic("jajsdj")
	return len(buf), nil
}

// bufReader reads from buffer
type BufferReader struct {
	buf [][]byte
	i   int
}

func NewBufferReader(c [][]byte) *BufferReader {
	return &BufferReader{buf: c}
}

func (r *BufferReader) Read(buf []byte) (int, error) {
	if len(r.buf) == r.i {
		return 0, io.EOF
	}
	buf = r.buf[r.i]
	r.i++
	return len(buf), nil
}

func (r *BufferReader) Close() error {
	return nil
}
