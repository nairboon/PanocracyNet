package main

import (
	"bitbucket.org/kardianos/osext"
	"fmt"
	flag "github.com/ogier/pflag"

	"log"
	"os"
	"os/exec"
	"path"
	"os/signal"
)

var n int
var dir string

func main() {
	flag.IntVar(&n, "peers", 2, "how many peers")
	flag.Parse()

	log.Printf("staring smulation with %d peers\n", n)
	rpcport := 9000
	p2pport := 10000
	peers := make(map[int]*exec.Cmd)

for i := 0; i < n; i++ {
        rpcportflag := fmt.Sprintf("--rpc-port=%d", rpcport + i)
        p2pportflag := fmt.Sprintf("--p2p-port=%d", p2pport + i)
	dirflag := fmt.Sprintf("--dir=peer%d", i)
		idflag := fmt.Sprintf("--id=peerID%d", i)
	filename, _ := osext.Executable()
	wd := path.Dir(filename)

	peer := exec.Command(wd+"/../bin/launcher", rpcportflag, p2pportflag, dirflag,idflag)
	peer.Stdout = os.Stdout
	peer.Stderr = os.Stderr
	err := peer.Start()
	if err != nil {
		log.Fatal(err)
	}
	peers[i] = peer

}

	





// wait for SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for _ = range c {
		break
	}
	log.Printf("exiting: ",)
	for i, p := range peers {
		p.Process.Signal(os.Interrupt)
		log.Printf("stopping %d...", i)
		p.Process.Kill()

	}
}
