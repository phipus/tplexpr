package tplexpr

import (
	"encoding/json"
	"fmt"
)

type Args struct {
	args []Value
}

func (c *Args) Len() int {
	return len(c.args)
}

func (c *Args) GetDefault(i int, def Value) (v Value) {
	if i >= 0 && i < len(c.args) {
		v = c.args[i]
	} else {
		v = def
	}
	return
}

func (c *Args) Get(i int) (v Value) {
	if i >= 0 && i < len(c.args) {
		return c.args[i]
	}
	return Nil
}

func (c *Args) All() []Value {
	return append([]Value(nil), c.args...)
}

func BuiltinList(args Args) (Value, error) {
	return ListValue(args.All()), nil
}

type rangeIter struct {
	start int
	stop  int
	step  int
}

var _ ValueIter = &rangeIter{}

func (r *rangeIter) Next() (Value, error) {
	if (r.step > 0 && r.start >= r.stop) || (r.step < 0 && r.start <= r.stop) {
		return nil, ErrIterExhausted
	}
	v := NumberValue(r.start)
	r.start += r.step
	return v, nil
}

func BuiltinRange(args Args) (Value, error) {
	rng := &rangeIter{step: 1}

	switch args.Len() {
	case 0:
		// nop
	case 1:
		stop, err := args.Get(0).Number()
		if err != nil {
			return nil, err
		}
		rng.stop = int(stop)
	case 2:
		start, err := args.Get(0).Number()
		if err != nil {
			return nil, err
		}
		stop, err := args.Get(1).Number()
		if err != nil {
			return nil, err
		}
		rng.start = int(start)
		rng.stop = int(stop)
	default:
		start, err := args.Get(0).Number()
		if err != nil {
			return nil, err
		}
		stop, err := args.Get(1).Number()
		if err != nil {
			return nil, err
		}
		step, err := args.Get(2).Number()
		if err != nil {
			return nil, err
		}
		rng.start = int(start)
		rng.stop = int(stop)
		rng.step = int(step)
	}

	return IterValue{rng}, nil
}

func BuiltinGet(args Args) (Value, error) {
	obj, err := args.Get(0).Object()
	if err != nil {
		return nil, err
	}
	key, err := args.Get(1).String()
	if err != nil {
		return nil, err
	}
	value, ok := obj.Key(key)
	if !ok {
		value = args.Get(2)
	}
	return value, nil
}

func valueToJSON(v Value, t *interface{}) error {
	switch v.Kind() {
	case KindNil:
		*t = nil
		return nil
	case KindBool:
		*t = v.Bool()
		return nil
	case KindNumber:
		nr, err := v.Number()
		if err != nil {
			return err
		}
		*t = nr
		return nil
	case KindString:
		s, err := v.String()
		if err != nil {
			return err
		}
		*t = s
		return nil
	case KindList, KindIterator:
		lst, err := v.List()
		if err != nil {
			return err
		}
		tl := make([]interface{}, len(lst))
		for i := range lst {
			err = valueToJSON(lst[i], &tl[i])
			if err != nil {
				return err
			}
		}
		*t = tl
		return nil
	case KindObject:
		obj, err := v.Object()
		if err != nil {
			return err
		}
		to := make(map[string]interface{})
		for _, key := range obj.Keys() {
			v, ok := obj.Key(key)
			if ok {
				var t interface{}
				err = valueToJSON(v, &t)
				if err != nil {
					return err
				}
				to[key] = t
			}
		}
		*t = to
		return nil
	case KindFunction:
		r, err := Call(v, nil)
		if err != nil {
			return err
		}
		return valueToJSON(r, t)
	default:
		*t = fmt.Sprintf("%v", v)
		return nil
	}
}

func BuiltinJSON(args Args) (Value, error) {
	var t interface{}
	err := valueToJSON(args.Get(0), &t)
	if err != nil {
		return nil, err
	}
	d, err := json.Marshal(t)
	return StringValue(d), err
}

func BuiltinKind(args Args) (Value, error) {
	return StringValue(args.Get(0).Kind().String()), nil
}

var baseBuiltins = map[string]Value{
	"list":  FuncValue(BuiltinList),
	"true":  True,
	"false": False,
	"nil":   Nil,
	"range": FuncValue(BuiltinRange),
	"get":   FuncValue(BuiltinGet),
	"json":  FuncValue(BuiltinJSON),
	"kind":  FuncValue(BuiltinKind),
}

func AddBaseBuiltins(c *Context) {
	for name, value := range baseBuiltins {
		c.Declare(name, value)
	}
}

func AddBuiltins(c *Context) {
	AddBaseBuiltins(c)
	AddNumberBuiltins(c)
	AddListBuiltins(c)
	AddStringBuiltins(c)
	AddToBuiltins(c)
}
