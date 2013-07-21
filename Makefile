
GOPATH := $(CURDIR)/native

all: daemon launcher library modules

protocol: 
	thrift --gen go -o protocol protocol/p2p_meta.thrift
	thrift --gen go -o protocol protocol/anevonet_rpc.thrift

daemon:
	gd -e native/src/daemon -o native/bin/daemon

launcher: 
	gd -e native/src/launcher -o native/bin/launcher

library:
	#mkdir -p native/goroot/src/github.com/nairboon/anevonet
	#ln -sf ../../../../../../native/src/lib native/goroot/src/github.com/nairboon/anevonet/
	gd -e native/src/libanevonet/
modules:
	mkdir -p native/bin/modules
	gd -e -I native/src/libanevonet native/src/modules/newscast -o native/bin/modules/newscast

.PHONY: protocol
