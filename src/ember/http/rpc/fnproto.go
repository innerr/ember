package rpc

import (
	"fmt"
	"reflect"
	"strings"
)

func (p FnProto) String() string {
	return fmt.Sprintf("%s%v => %v",
		p.Name,
		types(p.ArgNames, p.ArgTypes, "(", ")"),
		types(p.ReturnNames, p.ReturnTypes, "(", ")"))
}

func types(names []string, types []string, lb, rb string) string {
	str := lb
	for i, name := range names {
		str += types[i] + " " + name
		if i + 1 != len(names) {
			str += ", "
		}
	}
	return str + rb
}

func NewFnProto(fn reflect.Value, tag reflect.StructTag, name string) (p FnProto) {
	p = FnProto{Name: name}
	typ := fn.Elem().Type()
	for i := 0; i < typ.NumIn(); i++ {
		p.ArgTypes = append(p.ArgTypes, tname(typ.In(i)))
	}
	for i := 0; i < typ.NumOut(); i++ {
		p.ReturnTypes = append(p.ReturnTypes, tname(typ.Out(i)))
	}

	args := tag.Get("args")
	if len(args) == 0 {
		p.ArgNames = []string{}
	} else {
		for _, it := range strings.Split(args, ",") {
			p.ArgNames = append(p.ArgNames, strings.TrimSpace(it))
		}
	}

	ret := tag.Get("return")
	if len(ret) == 0 {
		p.ReturnNames = []string{}
	} else {
		for _, it := range strings.Split(ret, ",") {
			p.ReturnNames = append(p.ReturnNames, strings.TrimSpace(it))
		}
	}
	return
}

type FnProto struct {
	Name        string   "name"
	ArgNames    []string "arg_names"
	ArgTypes    []string "arg_types"
	ReturnNames []string "return_names"
	ReturnTypes []string "return_types"
}

func tname(typ reflect.Type) (tn string) {
	kind := typ.Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		tn = tname(typ.Elem()) + "[]"
		return
	}
	if kind == reflect.Map {
		tn = "Map<" + tname(typ.Key()) + ", " + tname(typ.Elem()) + ">"
		return
	}
	if kind == reflect.Ptr {
		tn = tname(typ.Elem())
		return
	}
	if kind == reflect.Struct {
		tn = sname(typ)
		return
	}
	tn = typ.Name()
	if tn != "" {
		return
	}
	tn = kind.String()
	return
}

func sname(typ reflect.Type) (sn string) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Name[0] < 'A' || field.Name[0] > 'Z' {
			continue
		}
		sn += tname(field.Type) + " " + field.Name + ", "
	}
	if len(sn) == 0 {
		return "{}"
	}
	return typ.Name() + "{" + sn[0: len(sn) - 2] + "}"
}
