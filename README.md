PanocracyNet
============

This is a hub for software projects, with the aim to decentralize and distribute our communication infrastucture.


###AnEvoNet
A self evolving p2p system that adaptively chooses different algorithms for different use cases. 
The network has self-healing capabilities: with the SLACER algorithm, the behaviour of freeriders would be copied by its connected peers, thus behave self-desctructing until mutual service collapses. After that the non-malicious peers would continue the last working strategy. 

##### What does work so far?
Clients using the Golang client library use local udp peer discovery and form a network. Currently the only algorithm in use is newscast gossip.

#### Architecture
"Clients"/modules connect to a local daemon through ZeroMQ and communicate according to a thrift rpc-protocol. They can export their own rpc-api as a thrift protocol, which is then tunneled over the local daemon to a remote daemon.

[Read More](https://github.com/nairboon/PanocracyNet/blob/develop/AnEvoNet/README.md)

