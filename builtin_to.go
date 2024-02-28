package tplexpr

func BuiltinToBool(args Args) (Value, error) {
	value := args.Get(0)
	return BoolValue(value.Bool()), nil
}

func BuiltinToNumber(args Args) (Value, error) {
	nr, err := args.Get(0).Number()
	return NumberValue(nr), err
}

func BuiltinToString(args Args) (Value, error) {
	s, err := args.Get(0).String()
	return StringValue(s), err
}

func BuiltinToList(args Args) (Value, error) {
	l, err := args.Get(0).List()
	return ListValue(l), err
}

func BuiltinToObject(args Args) (Value, error) {
	o, err := args.Get(0).Object()
	return objectMapper{o}, err
}

var toBuiltins = map[string]Value{
	"toBool":   FuncValue(BuiltinToBool),
	"toNumber": FuncValue(BuiltinToNumber),
	"toString": FuncValue(BuiltinToString),
	"toList":   FuncValue(BuiltinToList),
	"toObject": FuncValue(BuiltinToObject),
}

func AddToBuiltins(c *Context) {
	for name, value := range toBuiltins {
		c.Declare(name, value)
	}
}
