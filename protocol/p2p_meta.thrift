namespace go external_rpc

include "Common.thrift"

struct News {
  1: string version = Common.VERSION,
  2: string node_id,
  3: string tag,
  4: map<string,Common.DNA> strategies,
  5: map<string,double> strategy_success,
 6: Common.Gene ja
}

service remote_rpc {
  void ping(),
  News gossip(1:News h),
  oneway void bye()
}
