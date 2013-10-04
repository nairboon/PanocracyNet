namespace go p2p_rpc

include "Common.thrift"

struct Hello {
  1: string Version = Common.VERSION,
  2: string NodeID,
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
  Hello hi(1:Hello h),
  #News gossip(1:News h),
  #oneway void bye()
}
