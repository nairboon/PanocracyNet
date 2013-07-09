namespace js services


include "common.thrift"


struct Hello {
  1: string version = common.VERSION,
  2: string node_id,
  3: string tag,
  4: map<string,double> strategy_success
}

service AnEvoNet {
  void ping(),
  Hello greet(1:Hello h),
  oneway void bye()
}
