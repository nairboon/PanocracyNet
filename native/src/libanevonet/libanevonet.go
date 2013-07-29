package libanevonet

import (
	flag "github.com/ogier/pflag"
	"log"
	"Common"
)

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
	c.Connect(port)

 return c
}
