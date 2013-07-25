
#GOPATH := $(CURDIR)

all: protocol daemon launcher library modules

protocol: 
	mkdir -p protocol
	thrift --gen go -r -out protocol ../../protocol/p2p_meta.thrift
	thrift --gen go -r -out protocol ../../protocol/anevonet_rpc.thrift
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
