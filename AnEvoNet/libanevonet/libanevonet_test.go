package libanevonet

import (
	"Common"
	"testing" //import go package for testing related functionality
)

func Test_ConnectToDaemon(t *testing.T) {
	_, err := NewModule("TestModule")
	if err != nil {
		t.Error("could not connect to daemon")
	}
}

func Test_NewModule(t *testing.T) {

	c, _ := NewModule("TestModule")
	var dna1 Common.P2PDNA
	var dna2 *Common.P2PDNA

	socket, err := c.Register(dna1, dna2)
	if err != nil || socket == "" {
		t.Error("could not register")
	}
}
