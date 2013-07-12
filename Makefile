

all: daemon launcher library modules

protocol: 
	thrift --gen go -o protocol protocol/p2p_meta.thrift
	thrift --gen go -o protocol protocol/anevonet_rpc.thrift

daemon:
	gd native/src/daemon -o native/bin/daemon

launcher: 
	gd native/src/launcher -o native/bin/launcher

library:
	#mkdir -p native/goroot/src/github.com/nairboon/anevonet
	#ln -sf ../../../../../../native/src/lib native/goroot/src/github.com/nairboon/anevonet/
	GOPATH=native gd native/src/libanevonet/
modules:
	mkdir -p native/bin/modules
	GOPATH=native gd -I native/src/libanevonet native/src/modules/newscast
	#GOPATH=native/goroot gd native/src/modules/newscast -o native/bin/modules/newscast

.PHONY: protocol
