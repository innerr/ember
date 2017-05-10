package main

import (
	"errors"
	"os"
	"ember/cli"
)

func main() {
	hub := cli.NewRpcHub(os.Args[1:], NewServer, &Client{}, "/")
	err := hub.Run()
	cli.Check(err)
}

type Client struct {
	Echo func(msg string) (echo string, err error) `args:"msg" return:"echo"`
	Panic func() (err error)
	Error func() (err error)
	Foo func(key string) (ret [][][]string, err error) `args:"key" return:"ret"`
}

func (p *Server) Echo(msg string) (echo string, err error) {
	echo = msg
	return
}

func (p *Server) Panic() (err error) {
	panic("panic as expected")
	return
}

func (p *Server) Error() (err error) {
	err = errors.New("error as expected")
	return
}

func NewServer(args []string) (p interface{}, err error) {
	p = &Server{}
	return
}

func (p *Server) Foo(key string) (ret [][][]string, err error) {
	ret = [][][]string{{{"foo"}}}
	return
}

type Server struct{}
