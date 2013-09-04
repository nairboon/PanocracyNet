namespace go anevonet_rpc

include "Common.thrift"

struct Module {
  1: string Name,
  2: Common.DNA DNA,
}

struct ConnectionReq {
 1: Common.Peer Target,
}

struct ConnectionRes {
 1: string Socket,
}

struct RegisterRes {
 1: string Socket,
 2: Common.DNA DNA,
}

service InternalRpc {
  RegisterRes RegisterModule(1:Module m),
  ConnectionRes RequestConnection(1:ConnectionReq m),
}
