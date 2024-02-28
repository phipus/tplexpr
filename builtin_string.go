package tplexpr

import "strings"

func BuiltinUpper(args Args) (Value, error) {
	s, err := args.Get(0).String()
	if err != nil {
		return nil, err
	}
	return StringValue(strings.ToUpper(s)), nil
}

func BuiltinLower(args Args) (Value, error) {
	s, err := args.Get(0).String()
	if err != nil {
		return nil, err
	}
	return StringValue(strings.ToLower(s)), nil
}

func AddStringBuiltins(c *Context) {
	c.Declare("upper", FuncValue(BuiltinUpper))
	c.Declare("lower", FuncValue(BuiltinLower))
}
