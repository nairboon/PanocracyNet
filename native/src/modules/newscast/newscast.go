package main


import (
 //ae "github.com/nairboon/anevonet/lib"
ae "libanevonet"
"log"
 //"Common"
"newscast_protocol"
)



func main() {
	log.Printf("newscasting")

con := ae.NewConnection()

dna := con.Register("Newscast", newscast_protocol.ROOT_DNA)
  /* connect to daemon
   register protocol
   name, dna 
   setup handlers
 
 	for 
	DNA change
	incoming msg
   start out msg loop
*/

}
