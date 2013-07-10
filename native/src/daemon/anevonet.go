package main

import "fmt"


type APICallChannel chan string

func main() {
	fmt.Printf("Starting anevonet daemon\n")
// declare channels
api_calls := make(APICallChannel, 5)

// read config file
// open database

// start evolver
// start external listener
// start connection manager
// start internal broker

	fmt.Printf("stopping anevonet daemon\n")
}
