namespace js services


const string VERSION = "0.1.0"


enum Transport {
 TCP, // native client <-> native client
 WEBSOCKET, // native client <-> browser
 WEBRTC, // browser <-> browser
 SCTP // future :)
}



