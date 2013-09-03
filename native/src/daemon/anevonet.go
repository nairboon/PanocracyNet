package main

import (
	log "github.com/golang/glog"
	//flag //"github.com/ogier/pflag"
	"flag"
)

type APICallChannel chan string

var port int
var dir string

func main() {
	flag.IntVar(&port, "port", 9000, "the port to start an instance of anevonet")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.Parse()

	log.Info("staring daemon on %d in %s\n", port, dir)
	// declare channels
	//api_calls := make(APICallChannel, 5)

	// read config file
	// open database

	// start evolver
	// start external listener
	// start connection manager
	// start internal broker

	log.Info("stopping anevonet daemon\n")
	log.Flush()
}
