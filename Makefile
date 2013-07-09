



protocol: 
	thrift --gen go -o protocol protocol/p2p_meta.thrift
	thrift --gen go -o protocol protocol/anevonet_rpc.thrift




.PHONY: protocol
