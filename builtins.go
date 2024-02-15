package tplexpr

import (
	"strings"
)

type Args struct {
	args []Value
}

func (c *Args) Argc() int {
	return len(c.args)
}

func (c *Args) ArgDefault(i int, def Value) (v Value) {
	if i >= 0 && i < len(c.args) {
		v = c.args[i]
	} else {
		v = def
	}
	return
}

var (
	EmptyStringValue Value = StringValue("")
	EmptyListValue   Value = ListValue{}
)

func (c *Args) Arg(i int) (v Value) {
	if i >= 0 && i < len(c.args) {
		return c.args[i]
	}
	return EmptyStringValue
}

func (c *Args) Args() []Value {
	return append([]Value(nil), c.args...)
}

func mapNOP(args Args) (Value, error) {
	return args.Arg(0), nil
}

func BuiltinMap(args Args) (Value, error) {
	values, err := args.ArgDefault(0, EmptyListValue).List()
	if err != nil {
		return nil, err
	}
	fn := args.ArgDefault(1, FuncValue(mapNOP))
	mapped := make(ListValue, len(values))
	for i, v := range values {
		mapped[i], err = Call(fn, []Value{v, NumberValue(i)})
		if err != nil {
			return nil, err
		}
	}
	return mapped, nil
}

func filterNOP(args Args) (Value, error) {
	return BoolValue(false), nil
}

func BuiltinFilter(args Args) (Value, error) {
	values, err := args.ArgDefault(0, EmptyListValue).List()
	if err != nil {
		return nil, err
	}
	fn := args.ArgDefault(1, FuncValue(filterNOP))

	filtered := []Value{}
	for _, value := range values {
		ok, err := Call(fn, []Value{value})
		if err != nil {
			return nil, err
		}
		if ok.Bool() {
			filtered = append(filtered, value)
		}
	}

	return ListValue(filtered), nil
}

func BuiltinReverse(args Args) (Value, error) {
	values, err := args.ArgDefault(0, EmptyListValue).List()
	if err != nil {
		return nil, err
	}

	reversed := make(ListValue, len(values))
	for i := 0; i < len(values); i++ {
		reversed[len(values)-i-1] = values[i]
	}

	return reversed, nil
}

func BuiltinJoin(args Args) (Value, error) {
	value, err := args.ArgDefault(0, EmptyListValue).List()
	if err != nil {
		return nil, err
	}
	sep, err := args.ArgDefault(1, StringValue("")).String()
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

func BuiltinBool(args Args) (Value, error) {
	value := args.Arg(0)
	return BoolValue(value.Bool()), nil
}

func BuiltinNumber(args Args) (Value, error) {
	value := args.Arg(0)
	n, err := value.Number()
	return NumberValue(n), err
}

func BuiltinList(args Args) (Value, error) {
	return ListValue(args.Args()), nil
}

func AddBuiltins(c *Context) {
	c.Declare("map", FuncValue(BuiltinMap))
	c.Declare("filter", FuncValue(BuiltinFilter))
	c.Declare("join", FuncValue(BuiltinJoin))
	c.Declare("list", FuncValue(BuiltinList))
	c.Declare("true", BoolValue(true))
	c.Declare("false", BoolValue(false))
	c.Declare("bool", FuncValue(BuiltinBool))
	c.Declare("number", FuncValue(BuiltinNumber))
}
