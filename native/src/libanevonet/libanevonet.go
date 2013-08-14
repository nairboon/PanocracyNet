package libanevonet

import (
	flag "github.com/ogier/pflag"
	"log"
	"Common"
	"fmt"
"io"
	zmq "github.com/vaughan0/go-zmq"
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

}


/* Server, used for messages from the daemon*/
type server struct {

}


/* we abstract this and export an AnEvo Connection which is bidirectional */
type AnEvoConnection struct {
 Client client
 Server server

}

func (*AnEvoConnection) Connect(port int) {
 log.Printf("Connecting on: %d",port)

}

func (*AnEvoConnection) Register(name string, dna Common.DNA) Common.DNA {
 return dna // check db for better dna
}

func NewConnection() *AnEvoConnection {
  c := &AnEvoConnection{}
  var port int
 	flag.IntVar(&port, "port", 9000, "port of the daemon")
	flag.Parse()



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
 
