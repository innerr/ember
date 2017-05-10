package main

import (
	"fmt"
	"os"
	"strconv"
	"ember/http/rpc"
	server "ember/demo/demo_simple/demo_simple_server"
)

func main() {
	err := run()
	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}

func run() (err error) {
	client := server.Client{}
	rpc := rpc.NewClient("127.0.0.1:" + strconv.Itoa(server.DemoBindingPort))

	// call api

	err = rpc.Reg(&client)
	if err != nil {
		return
	}

	fmt.Println("calling: Echo(hello world)")
	msg, err := client.Echo("hello world")
	if err != nil {
		return
	}
	fmt.Println("result:  " + msg)
	fmt.Println()

	fmt.Println("calling: Foo(hi)")
	ret, err := client.Foo("hi")
	if err != nil {
		return
	}
	fmt.Printf("result:  %v\n", ret)
	fmt.Println()

	// list api
	fmt.Println("calling: Builtin.List()")
	fns, err := rpc.Builtin.List()
	if err != nil {
		return
	}
	fmt.Println("result:  ")
	for _, fn := range fns {
		fmt.Println(fn.String())
	}
	fmt.Println()

	// check server uptime
	fmt.Println("calling: Builtin.Uptime()")
	start, dura, err := rpc.Builtin.Uptime()
	if err != nil {
		return
	}
	fmt.Println("result:  ")
	fmt.Printf("server start at %d, started %d sec\n", start, dura)
	fmt.Println()

	// get measure data
	fmt.Println("calling: Measure.Sync(0)")
	md, err := rpc.Measure.Sync(0)
	if err != nil {
		return
	}
	fmt.Println("result:  ")
	md.PrintReadable(40)
	fmt.Println()

	return
}
