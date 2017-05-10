package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"ember/http/rpc"
)

func (p *RpcHub) Run() (err error) {
	if len(p.args) == 0 {
		Errln("usage:\n  <bin> [-host=" + DefaultHost + "] [-port=" + DefaultPort + "] command [args]\n\ncommand:")
		p.cmds.Help(true)
		return ErrWrongArgs
	}
	p.cmds.Run(p.args)
	return
}

func (p *RpcHub) RpcClient() *rpc.Client {
	return p.client
}

func (p *RpcHub) Cmds() *Cmds {
	return p.cmds
}

func (p *RpcHub) Mux() *http.ServeMux {
	return p.mux
}

func (p *RpcHub) CmdRun(args []string) {
	sobj, err := p.sfunc(args)
	Check(err)
	rpc := rpc.NewServer()
	err = rpc.Reg(sobj, p.cobj)
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
	for _, fn := range fns {
		fmt.Printf("  " + fn.String())
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
	err = data.Dump(os.Stdout, true, 40)
	Check(err)
}

func NewRpcHub(args []string, sfunc NewServerFunc, cobj interface{}, path string) (p *RpcHub)  {
	host, args := PopArg("host", DefaultHost, args)
	portstr, args := PopArg("port", DefaultPort, args)
	port, err := strconv.Atoi(portstr)
	Check(err)

	addr := host + ":" + portstr + path

	client := rpc.NewClient(addr)
	err = client.Reg(cobj)
	Check(err)

	p = &RpcHub{host, port, args, NewCmds(), sfunc, cobj, client, http.NewServeMux(), path}

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
	sfunc NewServerFunc
	cobj interface{}
	client *rpc.Client
	mux *http.ServeMux
	path string
}

type NewServerFunc func(args []string)(interface{}, error)

const DefaultHost = "127.0.0.1"
const DefaultPort = "8080"
