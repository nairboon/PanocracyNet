namespace go p2p_rpc

include "Common.thrift"

struct HelloSYN {
  1: string Version = Common.VERSION,
  2: string NodeID,
}

struct HelloSYNACK {
  1: string Version = Common.VERSION,
  2: string NodeID,
  3: map<Common.Transport,i32> SupportedTransport
}

struct ConnectSYN {
  1: string NodeID,
}

struct Message {
	1: string Module,
	2: string Payload
}

struct News {
  1: string version = Common.VERSION,
  2: string node_id,
  3: string tag,
  4: map<string,Common.P2PDNA> strategies,
  5: map<string,double> strategy_success,
 6: Common.Gene ja
}

service remote_rpc {
  HelloSYNACK Hello(1:HelloSYN h),
  bool Connect(1:ConnectSYN c),

  Message SendMessage(1:Message m)

}
