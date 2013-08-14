namespace js services
namespace go Common

const string VERSION = "0.1.0"


enum Transport {
 TCP, // native client <-> native client
 WEBSOCKET, // native client <-> browser
 WEBRTC, // browser <-> browser
 SCTP // future :)
}


struct Gene {
  1: i32 value
}

typedef map<string,i32> DNA

struct Peer {
 1: string ip
}

struct Timestamp {
 1: i32 sec
}
