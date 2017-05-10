package rpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"ember/measure"
)

func (p *Server) Run(path string, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc(path, p.Serve)
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: mux}
	server.SetKeepAlivesEnabled(false)
	return server.ListenAndServe()
}

func (p *Server) Serve(w http.ResponseWriter, r *http.Request) {
	begin := time.Now().UnixNano()
	cost := begin
	cb := 0

	url := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	name := url[len(url) - 1]

	var err error
	defer func() {
		end := time.Now().UnixNano()
		p.measure.Record("api.microsec.intenal." + name, (cost - begin) / 1000)
		p.measure.Record("api.microsec.all." + name, (end - begin) / 1000)
		p.measure.Record("api.bytes." + name, int64(cb))
		if err != nil {
			p.measure.Record("api.error." + name, 0)
		}
	}()

	var data []byte
	data, err = p.handle(name, w, r)
	cost = time.Now().UnixNano()
	cb = len(data)

	h := w.Header()
	h.Set("Content-Type", "text/json")

	if err == nil {
		w.WriteHeader(HttpCodeOK)
	} else {
		w.WriteHeader(HttpCodeErr)
	}

	_, err = w.Write(data)
	_, err = w.Write([]byte("\n"))
}

func (p *Server) handle(name string, w http.ResponseWriter, r *http.Request) (data []byte, err error) {
	var status string
	var detail string

	result, err := p.call(name, w, r)
	if err == nil {
		status = StatusOK
	} else {
		status = StatusErr
		result = nil
		detail = NewErrRpcFailed(err).Error()
	}

	resp := NewResponse(status, detail, result)
	data, eom := json.Marshal(resp)
	if eom != nil {
		err = eom
		resp = NewResponse(StatusErr, NewErrRpcFailed(err).Error(), nil)
		data, eom = json.Marshal(resp)
		if eom != nil {
			err = eom
		}
	}
	return
}

func (p *Server) call(name string, w http.ResponseWriter, r *http.Request) (result map[string]interface{}, err error) {
	fn, ok1 := p.fns[name]
	fv, ok2 := p.fvs[name]
	if !ok1 || !ok2 {
		err = fmt.Errorf("api '%s' not found", name)
		return
	}

	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	ret, err := fv.Invoke(fn.proto, data)
	if err != nil {
		return
	}
	result = format(fn, ret)
	return
}

func format(fn *FnTrait, result []interface{}) (m map[string]interface{}) {
	retArgs := fn.proto.ReturnNames
	m = make(map[string]interface{})
	for index, value := range retArgs {
		if index < len(result) {
			m[value] = result[index]
		}
	}
	return
}

func (p *Server) GetMeasure() *measure.Measure {
	return p.measure
}

func (p *Server) Uptime() (start int64, dura int64, err error) {
	now := time.Now().UnixNano()
	start = p.start / int64(time.Second)
	dura = (now - p.start) / int64(time.Second)
	return
}

func (p *Server) List() (protos []FnProto, err error) {
	protos = p.fns.List()
	return
}

func (p *Server) Reg(sobj interface{}, cobj interface{}) (err error) {
	return p.reg("", sobj, cobj)
}

func (p *Server) reg(prefix string, sobj interface{}, cobj interface{}) (err error) {
	fns := NewFnTraits(cobj)
	fvs, err := NewFnValues(sobj, fns)
	if err != nil {
		return
	}
	for origin, fn := range fns {
		name := prefix + origin
		if _, ok := p.fns[name]; ok {
			err = fmt.Errorf("%s has registered", name)
			return
		}
		fn.Prefix(prefix)
		p.fns[name] = fn
		if _, ok := fvs[origin]; !ok {
			continue
		}
		p.fvs[name] = fvs[origin]
	}
	return
}

func NewServer() (p *Server) {
	return NewServerEx(true, true)
}

func NewServerEx(exportBuiltinApi bool, exportMeasureApi bool) (p *Server) {
	p = &Server {
		make(FnTraits),
		make(FnValues),
		measure.NewMeasure(time.Second * 60, time.Second * 60 * 60 * 24),
		time.Now().UnixNano(),
	}

	if exportMeasureApi {
		err := p.reg(MeasurePrefix, p.measure, &Measure{})
		if err != nil {
			panic(err)
		}
	}

	if exportBuiltinApi {
		err := p.reg(BuiltinPrefix, p, &Builtin{})
		if err != nil {
			panic(err)
		}
	}

	return
}

type Server struct {
	fns FnTraits
	fvs FnValues
	measure *measure.Measure
	start int64
}

func (p *ErrRpcFailed) Error() string {
	return "[rpc server] " + p.err.Error()
}

func NewErrRpcFailed(err error) error {
	if _, ok := err.(*ErrCallFailed); ok {
		return err
	}
	return &ErrRpcFailed{err}
}

type ErrRpcFailed struct {
	err error
}

func NewResponse(status, detail string, result map[string]interface{}) *Response {
	return &Response{status, detail, result}
}

type Response struct {
	Status string        `json:"status"`
	Detail string        `json:"detail"`
	Result map[string]interface{} `json:"result"`
}

const (
	StatusOK = "ok"
	StatusErr = "error"
	HttpCodeOK = http.StatusOK
	HttpCodeErr = 599
)
