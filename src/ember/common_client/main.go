package main

import (
	"fmt"
	"ember/http/rpc"
	"ember/cli"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	c := NewEmberClient(args, "/")
	if len(args) > 1 {
		c.Call(args)
	} else {
		c.List(args)
	}
}

func (p *EmberClient) List([]string) {
	fns, err := p.client.Builtin.List()
	cli.Check(err)
	p.help(fns)
}

func (p *EmberClient) Call(args []string) {
	if len(args) == 0 {
		return
	}
	if len(args) == 2 {
		err := p.client.SimpleCall(args[0], []string{})
		cli.Check(err)
		return
	}

	if len(args) > 2 {
		for _, arg := range args[1:] {
			if !strings.Contains(arg, "=") {
				fmt.Println("usage: <bin> func [arg1]=\"\" [arg2]=\"\" ...")
				break
			}
		}
	}

	err := p.client.SimpleCall(args[0], args[1:])
	cli.Check(err)
}

func (p *EmberClient) help(fns []rpc.FnProto) {
	types := func(names []string, types []string, lb, rb string) string {
		str := lb
		for i, name := range names {
			str += types[i] + " " + name
			if i + 1 != len(names) {
				str += ", "
			}
		}
		return str + rb
	}

	for _, fn := range fns {
		fmt.Printf("  %s%v => %v\n",
			fn.Name,
			types(fn.ArgNames, fn.ArgTypes, "(", ")"),
			types(fn.ReturnNames, fn.ReturnTypes, "(", ")"))
	}
}

func NewEmberClient(args []string, path string) (p *EmberClient) {
	host, args := cli.PopArg("host", "127.0.0.1", args)
	portStr, args := cli.PopArg("port", "", args)
	var addr string
	if portStr != "" {
		addr = host + path
	} else {
		addr = host + ":" + portStr + path
	}

	client := rpc.NewClient(addr)
	p = &EmberClient{client, path, args}
	return
}

type EmberClient struct {
	client *rpc.Client
	path string
	args []string
}
