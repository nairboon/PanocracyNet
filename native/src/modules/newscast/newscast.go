package main
/* Newscast Algorithm according to "Tag-Based Cooperation in Peer-to-Peer Networks with Newscast" (2005)

ACTIVE THREAD
while(TRUE) do
wait(!t);
neighbor = SELECTPEER();
SENDSTATE(neighbor);
n.state = RECEIVESTATE();
my.state.UPDATE(n.state);


PASSIVE THREAD
while(TRUE) do
n.state = RECEIVESTATE();
SENDSTATE(n.state.sender);
my.state.UPDATE(n.state);
*/


import (
 //ae "github.com/nairboon/anevonet/lib"
ae "libanevonet"
"log"
"time"
 "Common"
proto "newscast_protocol"
)



func main() {
	log.Printf("newscasting")

con := ae.NewConnection()

var dna Common.DNA
con.Register("Newscast", proto.RootDNA, dna)

mystate := &proto.PeerState{}
for con.ContinueRunning(dna) {
		time.Sleep( time.Duration(dna["sleep"]) * time.Millisecond)



}
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
