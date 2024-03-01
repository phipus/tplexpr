package tplexpr

import (
	"sort"
	"strings"
)

var listBuiltins = map[string]Value{
	"map":      FuncValue(BuiltinMap),
	"filter":   FuncValue(BuiltinFilter),
	"reversed": FuncValue(BuiltinReversed),
	"join":     FuncValue(BuiltinJoin),
	"append":   FuncValue(BuiltinAppend),
	"extend":   FuncValue(BuiltinExtend),
	"sorted":   FuncValue(BuiltinSorted),
	"min":      FuncValue(BuiltinMin),
	"max":      FuncValue(BuiltinMax),
	"reduce":   FuncValue(BuiltinReduce),
}

func AddListBuiltins(c *Context) {
	for name, value := range listBuiltins {
		c.Declare(name, value)
	}
}

var (
	mapSelf   = FuncValue(func(args Args) (Value, error) { return args.Get(0), nil })
	filterAny = FuncValue(func(args Args) (Value, error) { return True, nil })
)

type mapIter struct {
	src ValueIter
	fn  Value
	idx int
}

var _ ValueIter = &mapIter{}

func (i *mapIter) Next() (Value, error) {
	v, err := i.src.Next()
	if err == nil {
		v, err = Call(i.fn, []Value{v, NumberValue(i.idx)})
		i.idx += 1
	}
	return v, err
}

func BuiltinMap(args Args) (Value, error) {
	values, err := args.Get(0).Iter()
	if err != nil {
		return nil, err
	}
	fn := args.GetDefault(1, mapSelf)
	return IterValue{&mapIter{src: values, fn: fn}}, nil
}

type filterIter struct {
	src ValueIter
	fn  Value
}

var _ ValueIter = &filterIter{}

func (i *filterIter) Next() (Value, error) {
	for {
		v, err := i.src.Next()
		if err != nil {
			return nil, err
		}

		ok, err := Call(i.fn, []Value{v})
		if err != nil {
			return nil, err
		}

		if ok.Bool() {
			return v, nil
		}
	}
}

func BuiltinFilter(args Args) (Value, error) {
	values, err := args.Get(0).Iter()
	if err != nil {
		return nil, err
	}
	fn := args.GetDefault(1, filterAny)

	return IterValue{&filterIter{src: values, fn: fn}}, nil
}

func BuiltinReversed(args Args) (Value, error) {
	lst, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	reversed := make(ListValue, len(lst))
	for i := len(lst) - 1; i >= 0; i-- {
		reversed[len(lst)-i-1] = lst[i]
	}
	return reversed, nil
}

func BuiltinJoin(args Args) (Value, error) {
	value, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	sep, err := args.Get(1).String()
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

func BuiltinAppend(args Args) (Value, error) {
	lst, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	clone := make(ListValue, 0, len(lst)+args.Len()-1)
	clone = append(clone, lst...)
	for i := 1; i < args.Len(); i++ {
		clone = append(clone, args.Get(i))
	}
	return clone, nil
}

func BuiltinExtend(args Args) (Value, error) {
	lst, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	lst2, err := args.Get(1).List()
	if err != nil {
		return nil, err
	}

	clone := make(ListValue, 0, len(lst)+len(lst2))
	clone = append(clone, lst...)
	clone = append(clone, lst2...)

	return clone, nil
}

type sortableList struct {
	err error
	l   []Value
}

var _ sort.Interface = &sortableList{}

func (s *sortableList) Len() int {
	return len(s.l)
}

func (s *sortableList) Less(i, j int) (less bool) {
	if s.err != nil {
		return false
	}

	less, s.err = compareValues(s.l[i], s.l[j], LT)
	return
}

func (s *sortableList) Swap(i, j int) {
	s.l[i], s.l[j] = s.l[j], s.l[i]
}

func BuiltinSorted(args Args) (Value, error) {
	lst, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	reverse := args.Get(1).Bool()

	sorted := make([]Value, len(lst))
	copy(sorted, lst)

	s := sortableList{l: sorted}
	if reverse {
		sort.Sort(sort.Reverse(&s))
	} else {
		sort.Sort(&s)
	}
	return ListValue(sorted), s.err
}

func reduceIter(values ValueIter, fn func(Value, Value) (Value, error)) (Value, error) {
	value, err := values.Next()
	if err != nil {
		if err == ErrIterExhausted {
			return Nil, nil
		}
		return nil, err
	}

	for {
		next, err := values.Next()
		if err != nil {
			if err == ErrIterExhausted {
				err = nil
			}
			return value, err
		}

		value, err = fn(value, next)
		if err != nil {
			return value, err
		}
	}
}

func reduceArgs(args Args, fn func(Value, Value) (Value, error)) (v Value, err error) {
	var values ValueIter

	switch args.Len() {
	case 0:
		return Nil, nil
	case 1:
		values, err = args.Get(0).Iter()
		if err != nil {
			return
		}
	default:
		values = &listIter{args.All()}
	}

	return reduceIter(values, fn)
}

func reduceNumbers(fn func(n1, n2 float64) (Value, error)) func(v1, v2 Value) (Value, error) {
	return func(v1, v2 Value) (Value, error) {
		n1, err := v1.Number()
		if err != nil {
			return nil, err
		}
		n2, err := v2.Number()
		if err != nil {
			return nil, err
		}
		return fn(n1, n2)
	}
}

func BuiltinMax(args Args) (v Value, err error) {
	return reduceArgs(args, reduceNumbers(func(n1, n2 float64) (Value, error) {
		var max float64
		if n1 >= n2 {
			max = n1
		} else {
			max = n2
		}
		return NumberValue(max), nil
	}))
}

func BuiltinMin(args Args) (v Value, err error) {
	return reduceArgs(args, reduceNumbers(func(n1, n2 float64) (Value, error) {
		var max float64
		if n1 <= n2 {
			max = n1
		} else {
			max = n2
		}
		return NumberValue(max), nil
	}))
}

func BuiltinReduce(args Args) (v Value, err error) {
	values, err := args.Get(0).Iter()
	if err != nil {
		return nil, err
	}

	fn := args.Get(1)

	return reduceIter(values, func(v1, v2 Value) (Value, error) {
		return Call(fn, []Value{v1, v2})
	})
}
