package libzmqthrift

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "github.com/golang/glog"
	zmq "github.com/pebbe/zmq4"
	"github.com/samuel/go-thrift/thrift"
	"io"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
)

type RPCClient interface {
	Call(method string, request interface{}, response interface{}) error
}

const (
	Client zmq.Type = zmq.DEALER
	Server          = zmq.ROUTER
)

/* zeromq connection */
type ZmqConnection struct {
	Sock *zmq.Socket

	readbuf bytes.Buffer
	outbuf  bytes.Buffer
	mu      sync.Mutex
}

func (z *ZmqConnection) Read(buf []byte) (int, error) {
	//fmt.Println("zmq.READ")
	for {

		z.mu.Lock()
		msg, err := z.Sock.RecvMessage(zmq.DONTWAIT)

		if err == nil {
			id, _ := z.Sock.GetIdentity()
			fmt.Printf("%s, recieved %d parts: %s", id, len(msg), msg[0])
			for _, part := range msg {
				z.readbuf.WriteString(part)
			}
			/*for i := 0; i < len(msg); i++ {
				z.readbuf.WriteString(msg[i])
			}*/
		}
		// we have something for the reader
		if z.readbuf.Len() >= len(buf) {
			n, e := z.readbuf.Read(buf)
			z.mu.Unlock()
			return n, e
		}

		z.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

}
func (z *ZmqConnection) Write(buf []byte) (n int, err error) {
	z.mu.Lock()
	//fmt.Printf("'writing': %s\n", buf)
	n, err = z.outbuf.Write(buf)
	z.mu.Unlock()
	return
}
func (z *ZmqConnection) Send() {
	z.mu.Lock()
	fmt.Printf("sending: %v\n", z.outbuf.String())
	sent, err := z.Sock.SendMessage(z.outbuf.String())
	fmt.Printf("sent: %d %v\n", sent, err)
	z.outbuf.Reset()
	z.mu.Unlock()
	return
}

func (z *ZmqConnection) Close() error {
	fmt.Println("zmq.CLOSE")
	z.Sock.Close()
	return nil
}

type ThriftZMQChannel struct {
	*BufferReader
	//*ChanWriter
	*BufferWriter
}

type ChannelReadWriteCloser struct {
	*ChanReader
	*ChanWriter
}

func (t ThriftZMQChannel) Close() error {
	//t.R.Close()
	//fmt.Printf("closed tzmqc (%s)", t.Buf.String())
	return nil
}

func (z *ZmqConnection) NewThriftClient() RPCClient {
	client := thrift.NewClient(NewFramedReadWriteCloser(z, 0), thrift.NewBinaryProtocol(false, false), false)
	//client := thrift.NewClient(thrift.NewTransport(NewFramedReadWriteCloser(z, 0), thrift.BinaryProtocol), false)

	return client
}

func NewZMQConnection(name string, port int, t zmq.Type) (*ZmqConnection, error) {
	c := &ZmqConnection{}
	var err error

	c.Sock, err = zmq.NewSocket(t)
	if err != nil {
		return nil, err
	}
	if t == Server {
		c.Sock.SetIdentity("DAEMON")
		err = c.Sock.Bind(fmt.Sprintf("tcp://*:%d", port))
	} else {
		identity := fmt.Sprintf("%s -%04X", name, rand.Intn(0x10000))
		c.Sock.SetIdentity(identity)
		log.Infof("We are %s", identity)
		err = c.Sock.Connect(fmt.Sprintf("tcp://localhost:%d", port))
		if err != nil {
			return nil, err
		}
	}

	return c, err
}

// http://openvswitch.org/pipermail/dev/2010-November/004456.html
func FixUnixSocketPath(p string) string {
	if len(p) > 108 {

		dir := path.Dir(p)
		base := path.Base(p)

		f, _ := os.Open(dir)

		fd := f.Fd()

		np := fmt.Sprintf("/proc/self/fd/%d/%s", fd, base)
		return np
	} else {
		return p
	}

}
func NewZMQUnixConnection(path string) *ZmqConnection {
	c := &ZmqConnection{}
	var err error

	c.Sock, err = zmq.NewSocket(zmq.REQ)
	if err != nil {
		panic(err)
	}

	fp := FixUnixSocketPath(path)
	identity := fmt.Sprintf("%s -%04X", fp, rand.Intn(0x10000))
	c.Sock.SetIdentity(identity)

	log.Infof("We are %s on %s", identity, fp)
	err = c.Sock.Connect(fmt.Sprintf("ipc://%s", fp))
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	return c
}

type ChanReader struct {
	buf []byte
	c   <-chan []byte
}

func NewChanReader(c <-chan []byte) *ChanReader {
	return &ChanReader{c: c}
}

func (r *ChanReader) Read(buf []byte) (int, error) {
	for len(r.buf) == 0 {
		var ok bool
		r.buf, ok = <-r.c
		if !ok {
			return 0, io.EOF
		}
	}
	n := copy(buf, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

type ChanWriter struct {
	c chan<- []byte
}

func NewChanWriter(c chan<- []byte) *ChanWriter {
	return &ChanWriter{c: c}
}
func (w *ChanWriter) Write(buf []byte) (n int, err error) {
	b := make([]byte, len(buf))
	copy(b, buf)
	w.c <- b
	return len(buf), nil
}

// bufReader reads from buffer
type BufferReader struct {
	buf bytes.Buffer
}

func NewBufferReader(c []string) *BufferReader {
	b := &BufferReader{}
	for _, part := range c {
		b.buf.WriteString(part)
	}
	return b
}

func (r *BufferReader) Read(buf []byte) (int, error) {
	n, e := r.buf.Read(buf)
	//destsize := len(buf)
	//fmt.Printf("should read: %d br: %s\n", destsize, buf)

	return n, e
}

func (r *BufferReader) Close() error {
	//r.buf.Close()
	return nil
}
func (r *BufferReader) Send() {
	return
}

type BufferWriter struct {
	Buf bytes.Buffer
}

func NewBufferWriter() *BufferWriter {
	b := &BufferWriter{}

	return b
}

func (r *BufferWriter) Write(p []byte) (int, error) {
	n, e := r.Buf.Write(p)
	if e != nil {
		panic(e)
	}

	//fmt.Printf("wrote %d br: %s / %s\n", n, p, r.Buf.String())
	return n, e
}

func (r *BufferWriter) Close() error {
	//r.buf.Close()
	//fmt.Printf("buffer closed")
	return nil
}

type Sender interface {
	Send()
}

type ReadWriteSender interface {
	io.Reader
	io.Writer
	Sender
	io.Closer
}

/// framed for zmq

const (
	DefaultMaxFrameSize = 1024 * 1024
)

type ErrFrameTooBig struct {
	Size    int
	MaxSize int
}

func (e *ErrFrameTooBig) Error() string {
	return fmt.Sprintf("thrift: frame size while reading over allowed size (%d > %d)", e.Size, e.MaxSize)
}

type Flusher interface {
	Flush() error
}

type FramedReadWriteCloser struct {
	wrapped      ReadWriteSender
	maxFrameSize int
	rbuf         *bytes.Buffer
	wbuf         *bytes.Buffer
}

func NewFramedReadWriteCloser(wrapped ReadWriteSender, maxFrameSize int) *FramedReadWriteCloser {
	if maxFrameSize == 0 {
		maxFrameSize = DefaultMaxFrameSize
	}
	return &FramedReadWriteCloser{
		wrapped:      wrapped,
		maxFrameSize: maxFrameSize,
		rbuf:         &bytes.Buffer{},
		wbuf:         &bytes.Buffer{},
	}
}

func (f *FramedReadWriteCloser) Read(p []byte) (int, error) {
	if err := f.fillBuffer(); err != nil {
		return 0, err
	}

	i, e := f.rbuf.Read(p)
	//fmt.Printf("fread: %s\n", p)
	return i, e
}

func (f *FramedReadWriteCloser) ReadByte() (byte, error) {
	if err := f.fillBuffer(); err != nil {
		return 0, err
	}
	return f.rbuf.ReadByte()
}

func (f *FramedReadWriteCloser) fillBuffer() error {
	if f.rbuf.Len() > 0 {
		return nil
	}

	f.rbuf.Reset()
	frameSize := uint32(0)
	if err := binary.Read(f.wrapped, binary.BigEndian, &frameSize); err != nil {
		return err
	}
	if int(frameSize) > f.maxFrameSize {
		return &ErrFrameTooBig{int(frameSize), f.maxFrameSize}
	}
	// TODO: Copy may return the full frame and still return an error. In that
	//       case we could return the asked for bytes to the caller (and the error).
	if _, err := io.CopyN(f.rbuf, f.wrapped, int64(frameSize)); err != nil {
		return err
	}

	return nil
}

func (f *FramedReadWriteCloser) Write(p []byte) (int, error) {
	n, err := f.wbuf.Write(p)
	if err != nil {
		return n, err
	}
	if f.wbuf.Len() > f.maxFrameSize {
		return n, &ErrFrameTooBig{f.wbuf.Len(), f.maxFrameSize}
	}
	return n, nil
}

func (f *FramedReadWriteCloser) Close() error {
	return f.wrapped.Close()
}

func (f *FramedReadWriteCloser) Flush() error {
	frameSize := uint32(f.wbuf.Len())
	if frameSize > 0 {
		if err := binary.Write(f.wrapped, binary.BigEndian, frameSize); err != nil {
			return err
		}
		_, err := io.Copy(f.wrapped, f.wbuf)
		f.wrapped.Send()
		f.wbuf.Reset()
		return err
	}
	return nil
}
