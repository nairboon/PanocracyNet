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
"math/rand"
proto "newscast_protocol"
)


/*func updateState(mystate, newstate proto.PeerState) proto.PeerState {
// check if we need more peers
if len(mystate.NewsItems) < dna["cachesize"] {
	// we should 
}

}*/

func main() {
	log.Printf("newscasting")
r := rand.New(rand.NewSource(99))

con := ae.NewConnection()

var dna Common.DNA
con.Register("Newscast", proto.RootDNA, dna)

mystate := &proto.PeerState{}
for con.ContinueRunning(dna) {

// select Peer
 peer := mystate.NewsItems[r.Intn( len(mystate.NewsItems))].Agent
 pc := proto.NewscastClient{con.GetPeerConnection(peer)}




recstate, err := pc.ExchangeState(mystate)
	if err != nil {
		panic(err)
	}


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
