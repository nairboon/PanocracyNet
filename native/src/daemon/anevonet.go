package main

import (
	log "github.com/golang/glog"
	//flag //"github.com/ogier/pflag"
	"Common"
	irpc "anevonet_rpc"
	"errors"
	"flag"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
)

type Module struct {
	Name   string
	Socket string
	DNA    Common.DNA
}

type Anevonet struct {
	Engine  *xorm.Engine
	Modules map[string]*Module
	Dir     string
}

func (a *Anevonet) RegisterModule(module *irpc.Module) (*irpc.RegisterRes, error) {
	if _, ok := a.Modules[module.Name]; ok {
		return nil, errors.New("module already registered")
	}

	//TODO: fetch dna from db
	m := &Module{Name: module.Name, DNA: module.DNA, Socket: a.Dir + "/sockets/" + module.Name}
	a.Modules[module.Name] = m

	res := &irpc.RegisterRes{DNA: m.DNA, Socket: m.Socket}
	return res, nil
}

var port int
var dir string

func main() {
	flag.IntVar(&port, "port", 9000, "the port to start an instance of anevonet")
	flag.StringVar(&dir, "dir", "anevo", "working directory of anevonet")
	flag.Parse()

	log.Info("staring daemon on %d in %s\n", port, dir)

	d, _ := filepath.Abs(dir)
	a := Anevonet{Dir: d}
	engine, err := xorm.NewEngine("sqlite3", dir+"/anevonet.db")
	defer engine.Close()
	a.Engine = engine
	if err != nil {
		panic(err)
	}

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
