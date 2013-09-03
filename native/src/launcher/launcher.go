package main

import (
	"bitbucket.org/kardianos/osext"
	"fmt"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
)

var port int
var dir string

func main() {
	flag.IntVar(&port, "port", 9000, "the port to start an instance of anevonet")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.Parse()

	log.Printf("staring daemon on %d\n", port)
	portflag := fmt.Sprintf("--port=%d", port)
	dirflag := fmt.Sprintf("-dir=%s", dir)
	filename, _ := osext.Executable()
	wd := path.Dir(filename)

	daemon := exec.Command(wd+"/daemon", portflag, dirflag, "-log_dir='log'", "-alsologtostderr=true")
	daemon.Stdout = os.Stdout
	daemon.Stderr = os.Stderr
	err := daemon.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Daemon started...")

	modules := make(map[string]*exec.Cmd)

	modulesDir, err := ioutil.ReadDir(wd + "/modules")
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range modulesDir {
		c := exec.Command(wd+"/modules/"+m.Name(), portflag)
		err := c.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s started...", m.Name())
		modules[m.Name()] = c
	}

	err = daemon.Wait()
	log.Printf("Command finished with error: %v", err)
}
