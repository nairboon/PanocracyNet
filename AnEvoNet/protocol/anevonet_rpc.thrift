namespace go anevonet_rpc

include "Common.thrift"

struct Module {
  1: string Name,
  2: Common.P2PDNA DNA,
}

struct Service {
  1: string Name,
}

struct ConnectionReq {
 1: Common.Peer Target,
 2: string Module
}

struct ConnectionRes {
 1: string Socket,
}

struct RegisterRes {
 1: string Socket,
 2: Common.P2PDNA DNA,
3: Common.Peer ID
}

struct StatusRes {
1: Common.Peer ID, 
2: i32 OnlinePeers,
}

struct BootstrapRes {
 1: list<Common.Peer> Peers
}
struct BootstrapNetworkRes {
 1: bool Success
}

service InternalRpc {
StatusRes Status(),

	/* module management */
  RegisterRes RegisterModule(1:Module m),
  bool RegisterService(1:Service s),

/* connection management */
  ConnectionRes RequestConnection(1:ConnectionReq req),
  void ShutdownConnection(1:ConnectionRes req),

/* bootstrap management */
 /* to get initial peers for an algorithm */
  BootstrapRes BootstrapAlgorithm()
/*BootstrapNetworkRes*/ bool BootstrapNetwork(1:Common.Peer p)

}
