package grpc

import (
	"log"
	"testing"
)

type Name interface {
	Name() string
}

type A struct {
	a string
}

func (a A) Name() string {
	return "I'm a"
}

func (a A) Say() {
	log.Println(a.Name())
}

func (a A) SayReal(name Name) {
	log.Println(name.Name())
}

type B struct {
	A
}

func (b B) Name() string {
	return "I'm b"
}

type C struct {
	A
}

func (c C) Name() string {
	return "I'm c"
}

func TestHelloClient(t *testing.T) {
	// HelloClient()
	a := A{}
	b := B{}
	c := C{}
	a.Say()
	b.Say()
	c.Say()
	b.SayReal(b)
	c.SayReal(c)
}
