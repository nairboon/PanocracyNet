package main

/* Broadcast over UDP

(from anevonet.go)
6. local peer discovery with UDP broadcasts
6.1 Listen for other broadcasts and stop after we're connected
6.2 Broadcast this peer



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

	// 6.1 Listen for Broadcasts to bootstrap
	go func() {
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
				//log.Infof("asking to bootstrap with %d", res.Port)
				ok, err = con.Rpc.BootstrapNetwork(res)
				if !ok || err != nil {
					log.Errorln("bootstrapp didn't work", err)
				} else {
					log.Infoln("done", err)
					break
				}

			}
			time.Sleep(2500 * time.Millisecond)
		}
		log.Infoln("stop listening for broadcasts")
	}()

	// 6.2 Broadcast us
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
		time.Sleep(2000 * time.Millisecond)
	}

	fmt.Println("stopping discovery")

}
