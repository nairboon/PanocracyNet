package main

/* Broadcast over UDP

(from anevonet.go)
6. local peer discovery with UDP broadcasts
6.1 Broadcast this peer
6.2 Listen for other broadcasts


*/

import (
	"Common"
	"bytes"
	"fmt"
	log "github.com/golang/glog"
	"github.com/samuel/go-thrift/thrift"
	ae "libanevonet"
	"net"
	"time"
)

func main() {
	log.Info("discovering over udp")

	con := ae.NewConnection("UDPDiscovery")

	s, err := con.Rpc.Status()
	if err != nil {
		panic(err)
	}
	log.Info("got status")
	id := s.ID
	var ok bool
	// 6.1 Broadcast us
	go func() {
		c, err := net.ListenPacket("udp", ":0")
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		fmt.Println("Broadcasting this Peer...")
		for {

			dst, err := net.ResolveUDPAddr("udp", "255.255.255.255:8032")
			if err != nil {
				log.Fatal(err)
			}

			buf := &bytes.Buffer{}

			err = thrift.EncodeStruct(buf, thrift.NewBinaryProtocol(true, false), id)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := c.WriteTo(buf.Bytes(), dst); err != nil {
				log.Fatal(err)
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	// 6.2 Listen for Broadcasts
	addr := &net.UDPAddr{Port: 8032, IP: net.IPv4bcast}
	c, err := net.ListenUDP("udp4", addr)

	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()
	fmt.Println("listen for Broadcast...")
	for {

		b := make([]byte, 512)
		n, _, err := c.ReadFrom(b)
		if err != nil {
			log.Fatal(err)
		}
		b = b[:n]
		//log.Infoln(n, "bytes read from", peer)
		//log.Infoln(b)
		res := &Common.Peer{}
		buf := bytes.NewBuffer(b)
		err = thrift.DecodeStruct(buf, thrift.NewBinaryProtocol(true, false), res)
		if err != nil {
			log.Fatal(err)
		}
		if res.Port == id.Port {
			//log.Infoln("its us ...")
		} else {
			log.Infof("asking to bootstrap with %d", res.Port)
			ok, err = con.Rpc.BootstrapNetwork(res)
			if !ok {
				log.Errorln("bootstrapp didn't work", err)
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
	fmt.Println("stopping discovery")
}
