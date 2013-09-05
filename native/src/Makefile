
#GOPATH := $(CURDIR)

all: protocol daemon launcher library modules

protocol: 
	mkdir -p protocol
	rm -f protocol/p2p_meta.go
	rm -f protocol/anevonet_rpc.go
	rm -f protocol/Common.go
	rm -f protocol/anevonet_rpc.go_
	generator ../../protocol/p2p_meta.thrift protocol/p2p_meta.go
	generator -go.import Common ../../protocol/anevonet_rpc.thrift protocol/anevonet_rpc.go_
	sed 's/*Peer/*Common.Peer/' protocol/anevonet_rpc.go_ > protocol/anevonet_rpc.go
	rm -f protocol/anevonet_rpc.go_
	generator ../../protocol/Common.thrift protocol/Common.go
	#thrift --gen go -r -out protocol ../../protocol/p2p_meta.thrift
	#thrift --gen go -r -out protocol ../../protocol/anevonet_rpc.thrift
	gd protocol
daemon:
	mkdir -p ../bin
	gd -I protocol daemon -o ../bin/daemon
	cd daemon && go fmt

launcher: 
	gd launcher -o ../bin/launcher
	cd launcher && go fmt

library:
	gd -I protocol libanevonet/
	cd libanevonet && go fmt

modules:
	mkdir -p ../bin/modules
	$(MAKE) -C modules



.PHONY: protocol daemon launcher library modules
