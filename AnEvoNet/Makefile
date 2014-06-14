all: library daemon launcher modules simulation
gen = $(GOPATH)/bin/generator

protocol: 
	rm -f protocol/p2p_rpc.go
	rm -f protocol/p2p_rpc.go_
	rm -f protocol/anevonet_rpc.go
	rm -f protocol/Common.go
	rm -f protocol/anevonet_rpc.go_
	$(gen) -go.import Common protocol/p2p_rpc.thrift protocol/p2p_rpc.go_
	sed 's/*Peer/*Common.Peer/' protocol/p2p_rpc.go_ | sed 's/map\[string\]int32/*Common.P2PDNA/' | sed 's/map\[Transport\]int32/map\[Common.Transport\]int32/'   > protocol/p2p_rpc.go
	$(gen) -go.import Common protocol/anevonet_rpc.thrift protocol/anevonet_rpc.go_
	sed 's/*Peer/*Common.Peer/' protocol/anevonet_rpc.go_ | sed 's/map\[string\]int32/*Common.P2PDNA/'  > protocol/anevonet_rpc.go
	rm -f protocol/anevonet_rpc.go_
	rm -f protocol/p2p_rpc.go_
	$(gen) protocol/Common.thrift protocol/Common.go
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
	
dependencies:
	go get github.com/golang/glog
	go get github.com/lunny/xorm
	go get bitbucket.org/kardianos/osext
	go get github.com/ogier/pflag
	go get github.com/mattn/go-sqlite3
	go get github.com/pebbe/zmq4
	go get github.com/nairboon/go-thrift/thrift
	go get github.com/samuel/go-parser/parser
	git clone https://github.com/nairboon/go-thrift.git $(GOPATH)/src/github.com/samuel/go-thrift
	go install github.com/samuel/go-thrift/generator
	

.PHONY: protocol daemon launcher library modules simulation test dependencies
