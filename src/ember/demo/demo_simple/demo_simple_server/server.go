package server

import (
	"errors"
	"os"
	"strconv"
	"net/http"
	"ember/http/rpc"
)

func Main() {
	err := run(DemoBindingPort)
	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}

func run(port int) (err error) {
	server := NewServer()

	rpc := rpc.NewServer()
	err = rpc.Reg(server, &Client{})
	if err != nil {
		return
	}

	// both way works:

	return rpc.Run("/", port)

	http.HandleFunc("/", rpc.Serve)
	return http.ListenAndServe(":" + strconv.Itoa(port), nil)
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

func NewServer() (p *Server) {
	p = &Server{}
	return
}

func (p *Server) Foo(key string) (ret [][][]string, err error) {
	ret = [][][]string{{{"you are a fool"}}}
	return
}

type Server struct{}

const DemoBindingPort = 8899
