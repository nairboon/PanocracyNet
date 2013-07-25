namespace js services


include "Common.thrift"


struct Hello {
  1: string version = Common.VERSION,
  2: string node_id,
  3: string tag,
  4: map<string,double> strategy_success
}

service AnEvoNet {
  void ping(),
  Hello greet(1:Hello h),
  oneway void bye()
}
