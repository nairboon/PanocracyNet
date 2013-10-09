namespace js services
namespace go Common

const string VERSION = "0.1.0"

/*
 currently using tcp
*/
enum Transport {
 TCP, // native client <-> native client
 WEBSOCKET, // native client <-> browser
 WEBRTC, // browser <-> browser
 SCTP // future :), UDT, DCCP,µTP
}


struct Gene {
  1: i32 value
}

typedef map<string,i32> P2PDNA

struct Peer {
 1: string ID,
 2: string IP,
 3: i32 Port
}

struct Timestamp {
 1: i64 sec
}
