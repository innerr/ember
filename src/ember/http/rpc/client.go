package rpc

import (
	"bytes"
	"ember/measure"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
)

func (p *Client) Reg(obj interface{}) (err error) {
	return p.RegEx("", obj)
}

func (p *Client) List() (fns []FnProto) {
	return p.fns.List()
}

func (p *Client) SimpleCall(name string, args []string) (err error) {
	kvs := make(map[string]string)
	for _, arg := range args {
		kv := strings.Split(arg, "=")
		key := kv[0]
		value := kv[1]
		kvs[key] = value
	}

	postArgs, err := json.Marshal(kvs)
	if err != nil {
		return
	}
	resp, err := http.Post(p.addr+name, "text/json", bytes.NewReader(postArgs))
	if err != nil {
		return
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(string(data))
	return
}

func (p *Client) Call(name string, args []string) (ret []interface{}, err error) {
	fn := p.fns[name]
	if fn == nil {
		err = fmt.Errorf("%s not found", name)
		return
	}
	if len(args) != len(fn.proto.ArgTypes) {
		err = fmt.Errorf("%s args count not match", name)
		return
	}
	in := make([][]byte, len(args))
	for i, it := range args {
		if fn.proto.ArgTypes[i] == "string" {
			it = `"` + it + `"`
		}
		in[i] = []byte(it)
	}
	return fn.Call(in)
}

func (p *Client) RegEx(prefix string, obj interface{}) (err error) {
	fns := NewFnTraits(obj)
	for origin, fn := range fns {
		name := prefix + origin
		if _, ok := p.fns[name]; ok {
			err = fmt.Errorf("%s has registered", name)
			return
		}
		fn.Prefix(prefix)
		p.fns[name] = fn
		fn.Bind(p.addr + name)
	}
	return
}

func NewClient(addr string) (p *Client) {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	if !strings.HasSuffix(addr, "/") {
		addr = addr + "/"
	}

	p = &Client{addr: addr, fns: make(FnTraits)}

	err := p.RegEx(MeasurePrefix, &p.Measure)
	if err != nil {
		panic(err)
	}

	err = p.RegEx(BuiltinPrefix, &p.Builtin)
	if err != nil {
		panic(err)
	}
	return
}

type Client struct {
	addr string
	fns FnTraits
	Measure Measure
	Builtin Builtin
}

type Measure struct {
	Sync func(time int64) (measure.MeasureData, error) `args:"time" return:"data"`
}

type Builtin struct {
	List   func() ([]FnProto, error) `return:"protos"`
	Uptime func() (start int64, dura int64, err error) `return:"start,dura"`
}

var ErrApiExists = errors.New("api registered")

const MeasurePrefix = "Measure."
const BuiltinPrefix = "Builtin."
