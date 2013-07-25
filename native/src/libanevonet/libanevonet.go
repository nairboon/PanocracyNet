package libanevonet

import (
	flag "github.com/ogier/pflag"
	"log"
	"Common"
)

type AnEvoConnection struct {


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
