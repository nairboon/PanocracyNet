all: newscast udpdiscovery

newscast:
	rm -fR newscast/newscast_protocol
	mkdir -p newscast/newscast_protocol
	generator -go.import Common newscast/newscast.thirft newscast/newscast_protocol/newscast_protocol.go
	sed -i -r 's/Agent[ ]\*Peer/Agent \*Common.Peer/' newscast/newscast_protocol/newscast_protocol.go
	gd -I ../protocol -I ../libanevonet -I ../libzmqthrift -I newscast newscast -o ../../bin/modules/newscast
	cd newscast && go fmt

udpdiscovery:
	gd -I ../protocol -I ../libanevonet -I ../libzmqthrift udpdiscovery -o ../../bin/modules/udpdiscovery
	cd udpdiscovery && go fmt

.PHONY: newscast udpdiscovery
