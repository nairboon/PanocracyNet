
#GOPATH := $(CURDIR)

all: protocol daemon launcher library modules

protocol: 
	mkdir -p protocol
	rm -f protocol/p2p_meta.go
	rm -f protocol/anevonet_rpc.go
	rm -f protocol/Common.go
	generator ../../protocol/p2p_meta.thrift protocol/p2p_meta.go
	generator ../../protocol/anevonet_rpc.thrift protocol/anevonet_rpc.go
	generator ../../protocol/Common.thrift protocol/Common.go
	#thrift --gen go -r -out protocol ../../protocol/p2p_meta.thrift
	#thrift --gen go -r -out protocol ../../protocol/anevonet_rpc.thrift
	gd protocol
daemon:
	mkdir -p ../bin
	gd daemon -o ../bin/daemon

launcher: 
	gd launcher -o ../bin/launcher

library:
	gd -I protocol libanevonet/

modules:
	mkdir -p ../bin/modules
	$(MAKE) -C modules



.PHONY: protocol daemon launcher library modules
