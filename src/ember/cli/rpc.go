package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"ember/http/rpc"
)

func (p *RpcHub) Run() {
	if len(p.args) == 0 {
		Errln("usage:\n  <bin> [-host=" + DefaultHost + "] [-port=" + DefaultPort + "] command [args]\n\ncommand:")
		p.cmds.Help(true)
		os.Exit(1)
	}
	p.cmds.Run(p.args)
}

func (p *RpcHub) Cmds() *Cmds {
	return p.cmds
}

func (p *RpcHub) Mux() *http.ServeMux {
	return p.mux
}

func (p *RpcHub) CmdRun([]string) {
	sobj := p.sobj
	if reflect.TypeOf(sobj).Kind() == reflect.Func {
		out := reflect.ValueOf(sobj).Call([]reflect.Value{})
		err := rpc.IsError{out[len(out) - 1].Interface()}.Check()
		Check(err)
		sobj = out[0].Interface()
	}
	rpc := rpc.NewServer()
	err := rpc.Reg(sobj, p.cobj)
	Check(err)
	err = rpc.Run(p.path, p.port)
	Check(err)
}

func (p *RpcHub) CmdList([]string) {
	fns := p.client.List()
	p.help(fns)
}

func (p *RpcHub) CmdRemote([]string) {
	fns, err := p.client.Builtin.List()
	Check(err)
	p.help(fns)
}

func (p *RpcHub) help(fns []rpc.FnProto) {
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

func (p *RpcHub) CmdCall(args []string) {
	if len(args) == 0 {
		p.CmdList(args)
		return
	}
	ret, err := p.client.Call(args[0], args[1:])
	Check(err)

	if len(ret) == 0 || ret == nil {
		return
	}

	var obj interface{}
	obj = ret
	if len(ret) == 1 {
		obj = ret[0]
	}

	data, err := json.MarshalIndent(obj, "", "    ")
	Check(err)
	fmt.Println(string(data))
}

func (p *RpcHub) CmdStatus(args []string) {
	data, err := p.client.Measure.Sync(0)
	err = data.Dump(os.Stdout, true)
	Check(err)
}

func NewRpcHub(args []string, sobj interface{}, cobj interface{}, path string) (p *RpcHub)  {
	host, args := PopArg("host", DefaultHost, args)
	portstr, args := PopArg("port", DefaultPort, args)
	port, err := strconv.Atoi(portstr)
	Check(err)

	addr := host + ":" + portstr

	client := rpc.NewClient(addr)
	err = client.Reg(cobj)
	Check(err)

	p = &RpcHub{host, port, args, NewCmds(), sobj, cobj, client, http.NewServeMux(), path}

	p.cmds.Reg("run", "run server", p.CmdRun)
	p.cmds.Reg("list", "list api from local info", p.CmdList)
	p.cmds.Reg("remote", "list api from remote", p.CmdRemote)
	p.cmds.Reg("call", "call server api by: name [arg] [arg]...", p.CmdCall)
	p.cmds.Reg("status", "get server status", p.CmdStatus)
	return
}

type RpcHub struct {
	host string
	port int
	args []string
	cmds *Cmds
	sobj interface{}
	cobj interface{}
	client *rpc.Client
	mux *http.ServeMux
	path string
}

type NewServerFunc func()(interface{}, error)

const DefaultHost = "127.0.0.1"
const DefaultPort = "8080"