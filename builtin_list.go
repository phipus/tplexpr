package tplexpr

import (
	"sort"
	"strings"
)

var (
	mapSelf   = FuncValue(func(args Args) (Value, error) { return args.Get(0), nil })
	filterAny = FuncValue(func(args Args) (Value, error) { return True, nil })
)

func BuiltinMap(args Args) (Value, error) {
	values, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	fn := args.GetDefault(1, mapSelf)
	mapped := make(ListValue, len(values))
	for i, v := range values {
		mapped[i], err = Call(fn, []Value{v, NumberValue(i)})
		if err != nil {
			return nil, err
		}
	}
	return mapped, nil
}

func BuiltinFilter(args Args) (Value, error) {
	values, err := args.Get(0).List()
	if err != nil {
		return nil, err
	}
	fn := args.GetDefault(1, filterAny)

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

var (
	bMap      Value = FuncValue(BuiltinMap)
	bFilter   Value = FuncValue(BuiltinFilter)
	bReversed Value = FuncValue(BuiltinReversed)
	bJoin     Value = FuncValue(BuiltinJoin)
	bAppend   Value = FuncValue(BuiltinAppend)
	bExtend   Value = FuncValue(BuiltinExtend)
	bSorted   Value = FuncValue(BuiltinSorted)
)

func AddListBuiltins(c *Context) {
	c.Declare("map", bMap)
	c.Declare("filter", bFilter)
	c.Declare("reversed", bReversed)
	c.Declare("join", bJoin)
	c.Declare("append", bAppend)
	c.Declare("extend", bExtend)
	c.Declare("sorted", bSorted)
}
