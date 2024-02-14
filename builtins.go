package tplexpr

import (
	"fmt"
	"strings"
)

func GetArg(args []Value, i int, def Value) (v Value) {
	if i < len(args) {
		v = args[i]
	} else {
		v = def
	}
	return
}

func mapNOP(args []Value) (Value, error) {
	return GetArg(args, 0, StringValue("")), nil
}

func BuiltinMap(args []Value) (Value, error) {
	values, err := GetArg(args, 0, ListValue{}).List()
	if err != nil {
		return nil, err
	}
	fn := GetArg(args, 1, FuncValue(mapNOP))
	mapped := make(ListValue, len(values))
	for i, v := range values {
		mapped[i], err = fn.Call([]Value{v, StringValue(fmt.Sprint(v))})
		if err != nil {
			return nil, err
		}
	}
	return mapped, nil
}

func BuiltinJoin(args []Value) (Value, error) {
	value, err := GetArg(args, 0, ListValue{}).List()
	if err != nil {
		return nil, err
	}
	sep, err := GetArg(args, 1, StringValue("")).String()
	if err != nil {
		return nil, err
	}

	sb := strings.Builder{}
	for i, v := range value {
		if i != 0 {
			sb.WriteString(sep)
		}
		str, err := v.String()
		if err != nil {
			return nil, err
		}
		sb.WriteString(str)
	}

	return StringValue(sb.String()), nil
}

func BuiltinList(args []Value) (Value, error) {
	return ListValue(args), nil
}

func AddBuiltins(c *Context) {
	c.Declare("map", FuncValue(BuiltinMap))
	c.Declare("join", FuncValue(BuiltinJoin))
	c.Declare("list", FuncValue(BuiltinList))
}
