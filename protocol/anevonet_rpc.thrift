namespace js services
namespace go internal_rpc

include "Common.thrift"

struct Module {
  1: string name,
  2: Common.DNA dna,
}

service local_rpc {
  Common.DNA register_module(1:Module m),
}
