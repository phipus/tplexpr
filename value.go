package tplexpr

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ValueKind int

const (
	KindNil ValueKind = iota
	KindString
	KindBool
	KindNumber
	KindList
	KindObject
	KindFunction
	KindIterator
)

const (
	KindNilName      = "nil"
	KindStringName   = "string"
	KindBoolName     = "bool"
	KindNumberName   = "number"
	KindListName     = "list"
	KindObjectName   = "object"
	KindFunctionName = "function"
	KindIteratorName = "iterator"
)

func (v ValueKind) String() string {
	s := ""
	switch v {
	case KindString:
		s = KindStringName
	case KindBool:
		s = KindBoolName
	case KindNumber:
		s = KindNumberName
	case KindList:
		s = KindListName
	case KindObject:
		s = KindObjectName
	case KindFunction:
		s = KindFunctionName
	case KindIterator:
		s = KindIteratorName
	}
	return s
}

type Value interface {
	Kind() ValueKind
	Bool() bool
	Number() (float64, error)
	String() (string, error)
	List() ([]Value, error)
	Iter() (ValueIter, error)
	Object() (Object, error)
	Call(args Args, wr ValueWriter) error
}

type ValueIter interface {
	Next() (Value, bool, error)
}

type Object interface {
	Key(name string) (Value, bool)
	SetKey(name string, value Value)
	Keys() []string
}

type MapObject struct {
	M map[string]Value
}

func (m *MapObject) Key(name string) (v Value, ok bool) {
	v, ok = m.M[name]
	return
}

func (m *MapObject) SetKey(name string, value Value) {
	if m.M == nil {
		m.M = map[string]Value{}
	}
	m.M[name] = value
}

func (m *MapObject) Keys() []string {
	keys := make([]string, 0, len(m.M))
	for key := range m.M {
		keys = append(keys, key)
	}
	return keys
}

const (
	opConvert = "convert"
	opAdd     = "add"
	opSub     = "subtract"
	opMul     = "multiply"
	opDiv     = "divide"
	conTO     = "to"
	conBY     = "by"
	conOF     = "of"
)

type ErrType struct {
	Op   string
	From string
	con  string
	To   string
}

func (e *ErrType) Error() string {
	return fmt.Sprintf("type error: can not %s %s to %s", e.Op, e.From, e.To)
}

type nilValue struct{}

var Nil Value = nilValue{}

func (n nilValue) Kind() ValueKind {
	return KindNil
}

func (n nilValue) Bool() bool {
	return false
}

func (n nilValue) Number() (float64, error) {
	return 0, nil
}

func (n nilValue) String() (string, error) {
	return "", nil
}

func (n nilValue) List() ([]Value, error) {
	return nil, nil
}

func (n nilValue) Iter() (ValueIter, error) {
	return &listIter{}, nil
}

func (n nilValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (n nilValue) Call(args Args, wr ValueWriter) error {
	return nil
}

type BoolValue bool

var _ Value = BoolValue(false)

func (b BoolValue) Kind() ValueKind {
	return KindBool
}

func (b BoolValue) Bool() bool {
	return bool(b)
}

func (b BoolValue) Number() (float64, error) {
	if b {
		return 1, nil
	}
	return 0, nil
}

func (b BoolValue) String() (string, error) {
	return fmt.Sprint(bool(b)), nil
}

func (b BoolValue) List() ([]Value, error) {
	return []Value{b}, nil
}

func (b BoolValue) Iter() (ValueIter, error) {
	return &singleValueIter{b}, nil
}

func (b BoolValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (b BoolValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(b)
}

type NumberValue float64

func (n NumberValue) Kind() ValueKind {
	return KindNumber
}

func (n NumberValue) Bool() bool {
	return n != 0
}

func (n NumberValue) Number() (float64, error) {
	return float64(n), nil
}

func (n NumberValue) String() (string, error) {
	return fmt.Sprint(float64(n)), nil
}

func (n NumberValue) List() ([]Value, error) {
	return []Value{n}, nil
}

func (n NumberValue) Iter() (ValueIter, error) {
	return &singleValueIter{n}, nil
}

func (n NumberValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (n NumberValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(n)
}

type StringValue string

var _ Value = StringValue("")

func (s StringValue) Kind() ValueKind {
	return KindString
}

func (s StringValue) Bool() bool {
	return len(s) > 0
}

func (s StringValue) Number() (float64, error) {
	return strconv.ParseFloat(string(s), 64)
}

func (s StringValue) String() (string, error) {
	return string(s), nil
}

func (s StringValue) List() ([]Value, error) {
	return []Value{s}, nil
}

func (s StringValue) Iter() (ValueIter, error) {
	return &singleValueIter{s}, nil
}

func (s StringValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (s StringValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(s)
}

type ListValue []Value

var _ Value = ListValue{}

func (l ListValue) Kind() ValueKind {
	return KindList
}

func (l ListValue) Bool() bool {
	return len(l) > 0
}

func (l ListValue) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindListName, conTO, KindNumberName}
}

func (l ListValue) String() (string, error) {
	sb := strings.Builder{}
	for i, v := range l {
		if i != 0 {
			sb.WriteByte(' ')
		}
		str, err := v.String()
		if err != nil {
			return sb.String(), err
		}
		sb.WriteString(str)
	}
	return sb.String(), nil
}

func (l ListValue) List() ([]Value, error) {
	return l, nil
}

func (l ListValue) Iter() (ValueIter, error) {
	return &listIter{l}, nil
}

type ListObject struct {
	L []Value
}

func (l *ListObject) Key(name string) (Value, bool) {
	idx, err := strconv.ParseInt(name, 10, 64)
	if err == nil && idx < int64(len(l.L)) {
		return l.L[int(idx)], true
	}
	return nil, false
}

func (l *ListObject) SetKey(name string, value Value) {
	idx, err := strconv.ParseInt(name, 10, 64)
	if err == nil && idx < int64(len(l.L)) {
		l.L[int(idx)] = value
	}
}

func (l *ListObject) Keys() []string {
	keys := make([]string, len(l.L))
	for i := range l.L {
		keys[i] = fmt.Sprintf("%d", i)
	}
	return keys
}

func (l ListValue) Object() (Object, error) {
	return &ListObject{l}, nil
}

func (l ListValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(l)
}

type ObjectValue map[string]Value

var _ Value = ObjectValue{}

func (o ObjectValue) Kind() ValueKind {
	return KindObject
}

func (o ObjectValue) Bool() bool {
	return len(o) > 0
}

func (o ObjectValue) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindObjectName, conTO, KindNumberName}
}

func (o ObjectValue) String() (string, error) {
	sb := strings.Builder{}
	first := true
	for key := range o {
		if !first {
			sb.WriteByte(' ')
		} else {
			first = false
		}

		sb.WriteString(key)
	}

	return sb.String(), nil
}

func (o ObjectValue) List() ([]Value, error) {
	keys := make([]Value, 0, len(o))
	for key := range o {
		keys = append(keys, StringValue(key))
	}
	return keys, nil
}

func (o ObjectValue) Iter() (ValueIter, error) {
	keys, _ := o.List()
	return &listIter{keys}, nil
}

func (o ObjectValue) Object() (Object, error) {
	return &MapObject{o}, nil
}

func (o ObjectValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(o)
}

type FuncValue func(args Args) (Value, error)

var _ Value = FuncValue(nil)

func (f FuncValue) Kind() ValueKind {
	return KindFunction
}

func (f FuncValue) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindFunctionName, conTO, KindNumberName}
}

func (f FuncValue) Bool() bool {
	return true
}

func (f FuncValue) String() (string, error) {
	value, err := f(Args{})
	if err != nil {
		return "", err
	}
	return value.String()
}

func (f FuncValue) List() ([]Value, error) {
	value, err := f(Args{})
	if err != nil {
		return nil, err
	}
	return value.List()
}

func (f FuncValue) Iter() (ValueIter, error) {
	value, err := f(Args{})
	if err != nil {
		return nil, err
	}
	return value.Iter()
}

func (f FuncValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (f FuncValue) Call(args Args, wr ValueWriter) error {
	value, err := f(args)
	if err != nil {
		return err
	}
	return wr.WriteValue(value)
}

var IterListLimit = 10000

type IterValue struct {
	I ValueIter
}

var _ Value = IterValue{}

func (v IterValue) Kind() ValueKind {
	return KindIterator
}

func (v IterValue) Bool() bool {
	return true
}

func (v IterValue) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindIteratorName, conTO, KindNumberName}
}

func (v IterValue) String() (string, error) {
	sb := strings.Builder{}
	for i := 0; i < IterListLimit; i++ {
		v, ok, err := v.I.Next()
		if err != nil {
			return "", err
		}
		if !ok {
			return sb.String(), nil
		}

		if i != 0 {
			sb.WriteByte(' ')
		}

		str, err := v.String()
		if err != nil {
			return "", err
		}
		sb.WriteString(str)
	}
	sb.WriteString(" ...")
	return sb.String(), nil
}

var ErrIterListLimit = errors.New("iterator to list limit exhausted")

func (v IterValue) List() ([]Value, error) {
	lst := []Value{}
	for i := 0; i < IterListLimit; i++ {
		v, ok, err := v.I.Next()
		if err != nil {
			return lst, err
		}
		if !ok {
			return lst, nil
		}
		lst = append(lst, v)
	}
	return lst, ErrIterListLimit
}

func (v IterValue) Iter() (ValueIter, error) {
	return v.I, nil
}

func (v IterValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (v IterValue) Call(args Args, wr ValueWriter) error {
	vv, ok, err := v.I.Next()
	if err != nil {
		return err
	}
	if ok {
		return wr.WriteValue(vv)
	}
	return wr.WriteValue(Nil)
}

type subprogValue struct {
	code []Instr
	args []string
	ctx  *Context
}

var _ Value = &subprogValue{}

func (s *subprogValue) eval(args Args, wr ValueWriter) error {
	s.ctx.BeginScope()
	defer s.ctx.EndScope()

	for i, argName := range s.args {
		s.ctx.Declare(argName, args.Arg(i))
	}

	return EvalRaw(s.ctx, s.code, wr)
}

func (s *subprogValue) evalString(args Args) (string, error) {
	s.ctx.BeginScope()
	defer s.ctx.EndScope()

	for i, argName := range s.args {
		s.ctx.Declare(argName, args.Arg(i))
	}

	return EvalString(s.ctx, s.code)
}

func (s *subprogValue) Kind() ValueKind {
	return KindFunction
}

func (s *subprogValue) Bool() bool {
	return true
}

func (s *subprogValue) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindFunctionName, conTO, KindNumberName}
}

func (s *subprogValue) String() (string, error) {
	value, err := s.evalString(Args{})
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s *subprogValue) List() ([]Value, error) {
	value, err := s.evalString(Args{})
	if err != nil {
		return nil, err
	}
	return ListValue{StringValue(value)}, nil
}

func (s *subprogValue) Iter() (ValueIter, error) {
	value, err := s.evalString(Args{})
	if err != nil {
		return nil, err
	}
	return &singleValueIter{StringValue(value)}, nil
}

func (s *subprogValue) Object() (Object, error) {
	return &MapObject{}, nil
}

func (s *subprogValue) Call(args Args, wr ValueWriter) error {
	return s.eval(args, wr)
}

type objectMapper struct {
	o Object
}

var _ Value = objectMapper{}

func (o objectMapper) Kind() ValueKind {
	return KindObject
}

func (o objectMapper) Bool() bool {
	return true
}

func (o objectMapper) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindObjectName, conTO, KindNumberName}
}

func (o objectMapper) String() (string, error) {
	return strings.Join(o.o.Keys(), " "), nil
}

func (o objectMapper) List() ([]Value, error) {
	keys := o.o.Keys()
	values := make([]Value, len(keys))
	for i := range keys {
		values[i] = StringValue(keys[i])
	}
	return values, nil
}

func (o objectMapper) Iter() (ValueIter, error) {
	values, _ := o.List()
	return &listIter{values}, nil
}

func (o objectMapper) Object() (Object, error) {
	return o.o, nil
}

func (o objectMapper) Call(args Args, wr ValueWriter) error {
	wr.WriteValue(o)
	return nil
}
