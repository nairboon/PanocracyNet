// This file is automatically generated. Do not modify.

package Common

import (
	"strconv"
	"fmt"
)

type P2PDNA map[string]int32

const VERSION string = "0.1.0"

type Transport int32

var(
	TransportSCTP = Transport(3)
	TransportTCP = Transport(0)
	TransportWEBRTC = Transport(2)
	TransportWEBSOCKET = Transport(1)
	TransportByName = map[string]Transport{
		"Transport.SCTP": TransportSCTP,
		"Transport.TCP": TransportTCP,
		"Transport.WEBRTC": TransportWEBRTC,
		"Transport.WEBSOCKET": TransportWEBSOCKET,
	}
	TransportByValue = map[Transport]string{
		TransportSCTP: "Transport.SCTP",
		TransportTCP: "Transport.TCP",
		TransportWEBRTC: "Transport.WEBRTC",
		TransportWEBSOCKET: "Transport.WEBSOCKET",
	}
)

func (e Transport) String() string {
	name := TransportByValue[e]
	if name == "" {
		name = fmt.Sprintf("Unknown enum value Transport(%d)", e)
	}
	return name
}

func (e Transport) MarshalJSON() ([]byte, error) {
	name := TransportByValue[e]
	if name == "" {
		name = strconv.Itoa(int(e))
	}
	return []byte("\""+name+"\""), nil
}

func (e *Transport) UnmarshalJSON(b []byte) error {
	st := string(b)
	if st[0] == '"' {
		*e = Transport(TransportByName[st[1:len(st)-1]])
		return nil
	}
	i, err := strconv.Atoi(st)
	*e = Transport(i)
	return err
}

type Gene struct {
	Value int32 `thrift:"1,required" json:"value"`
}

type Peer struct {
	ID string `thrift:"1,required" json:"ID"`
	IP string `thrift:"2,required" json:"IP"`
	Port int32 `thrift:"3,required" json:"Port"`
}

type Timestamp struct {
	Sec int32 `thrift:"1,required" json:"sec"`
}