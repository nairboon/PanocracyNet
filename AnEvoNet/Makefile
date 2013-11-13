all: library daemon launcher modules

protocol: 
	rm -f protocol/p2p_rpc.go
	rm -f protocol/p2p_rpc.go_
	rm -f protocol/anevonet_rpc.go
	rm -f protocol/Common.go
	rm -f protocol/anevonet_rpc.go_
	generator -go.import Common protocol/p2p_rpc.thrift protocol/p2p_rpc.go_
	sed 's/*Peer/*Common.Peer/' protocol/p2p_rpc.go_ | sed 's/map\[string\]int32/*Common.P2PDNA/'  > protocol/p2p_rpc.go
	generator -go.import Common protocol/anevonet_rpc.thrift protocol/anevonet_rpc.go_
	sed 's/*Peer/*Common.Peer/' protocol/anevonet_rpc.go_ | sed 's/map\[string\]int32/*Common.P2PDNA/'  > protocol/anevonet_rpc.go
	rm -f protocol/anevonet_rpc.go_
	rm -f protocol/p2p_rpc.go_
	generator protocol/Common.thrift protocol/Common.go
	#thrift --gen go -r -out protocol ../../protocol/p2p_meta.thrift
	#thrift --gen go -r -out protocol ../../protocol/anevonet_rpc.thrift
	gd protocol
daemon:
	mkdir -p bin
	gd -I protocol -I libzmqthrift daemon -o bin/daemon
	cd daemon && go fmt

launcher: 
	gd launcher -o bin/launcher
	cd launcher && go fmt

library:
	gd libzmqthrift/
	cd libzmqthrift && go fmt
	gd -I protocol -I libzmqthrift libanevonet/
	cd libanevonet && go fmt

modules:
	mkdir -p bin/modules
	$(MAKE) -C modules

simulation:
	gd -o simulation/sim simulation

test:	daemon library
	-killall -q daemon
	./bin/daemon &
	gd -I protocol -I libzmqthrift -t libanevonet
	-killall -q daemon

.PHONY: protocol daemon launcher library modules simulation test
