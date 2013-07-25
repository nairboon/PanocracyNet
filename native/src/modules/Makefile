all: newscast

newscast:
	thrift --gen go --out newscast newscast/newscast.thirft
	gd -I ../protocol -I ../libanevonet newscast -o ../../bin/modules/newscast

.PHONY: newscast
