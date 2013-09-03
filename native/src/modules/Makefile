all: newscast

newscast:
	rm -fR newscast/newscast_protocol
	mkdir -p newscast/newscast_protocol
	generator -go.import Common newscast/newscast.thirft newscast/newscast_protocol/newscast_protocol.go
	sed -i -r 's/Agent[ ]\*Peer/Agent \*Common.Peer/' newscast/newscast_protocol/newscast_protocol.go
	gd -I ../protocol -I ../libanevonet -I newscast newscast -o ../../bin/modules/newscast
	cd newscast && go fmt
.PHONY: newscast
