package tplexpr

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ValueKind int

const (
	KindString ValueKind = iota
	KindBool
	KindNumber
	KindList
	KindObject
	KindFunction
	KindIterator
)

const (
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
	Keys() []string
	GetAttr(name string) (Value, bool)
	Call(args Args, wr ValueWriter) error
}

type ValueIter interface {
	Next() (Value, bool, error)
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

func (b BoolValue) Keys() []string {
	return nil
}

func (b BoolValue) GetAttr(name string) (Value, bool) {
	return nil, false
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

func (n NumberValue) Keys() []string {
	return nil
}

func (n NumberValue) GetAttr(name string) (Value, bool) {
	return nil, false
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

func (s StringValue) Keys() []string {
	return nil
}

func (s StringValue) GetAttr(name string) (Value, bool) {
	return nil, false
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

func (l ListValue) Keys() []string {
	keys := make([]string, len(l))
	for i := range l {
		keys[i] = fmt.Sprintf("%d", i)
	}
	return keys
}

func (l ListValue) GetAttr(name string) (Value, bool) {
	idx, err := strconv.ParseInt(name, 10, 64)
	if err != nil {
		return nil, false
	}

	if idx < int64(len(l)) {
		return l[int(idx)], true
	}

	return nil, false
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

func (o ObjectValue) Keys() []string {
	keys := make([]string, 0, len(o))
	for key := range o {
		keys = append(keys, key)
	}
	return keys
}

func (o ObjectValue) GetAttr(name string) (v Value, ok bool) {
	v, ok = o[name]
	return
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

func (f FuncValue) Keys() []string {
	return nil
}

func (f FuncValue) GetAttr(name string) (Value, bool) {
	return nil, false
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

func (v IterValue) Keys() []string {
	return nil
}

func (v IterValue) GetAttr(name string) (Value, bool) {
	return nil, false
}

func (v IterValue) Call(args Args, wr ValueWriter) error {
	vv, ok, err := v.I.Next()
	if err != nil {
		return err
	}
	if ok {
		return wr.WriteValue(vv)
	}
	return wr.WriteValue(EmptyStringValue)
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
		s.ctx.Declare(argName, args.ArgDefault(i, StringValue("")))
	}

	return EvalRaw(s.ctx, s.code, wr)
}

func (s *subprogValue) evalString(args Args) (string, error) {
	s.ctx.BeginScope()
	defer s.ctx.EndScope()

	for i, argName := range s.args {
		s.ctx.Declare(argName, args.ArgDefault(i, StringValue("")))
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

func (s *subprogValue) Keys() []string {
	return nil
}

func (s *subprogValue) GetAttr(name string) (Value, bool) {
	return nil, false
}

func (s *subprogValue) Call(args Args, wr ValueWriter) error {
	return s.eval(args, wr)
}

type Subprog struct {
	Args []string
	Code []Instr
}

type Context struct {
	vars          map[string]Value
	shadowed      []namedVar
	scope         int
	prevScopes    []int
	subprogs      []Subprog
	iters         []ValueIter
	valueFilters  []ValueFilter
	outputFilters []ValueFilter
	NameError     func(name string) (Value, error)
}

func NewContext() Context {
	return Context{
		vars: map[string]Value{},
	}
}

func (c *Context) Clone() *Context {
	clone := NewContext()
	for name, value := range c.vars {
		clone.vars[name] = value
	}
	clone.subprogs = c.subprogs
	clone.NameError = c.NameError
	return &clone
}

type namedVar struct {
	name  string
	value Value
}

func (c *Context) TryLookup(name string) (value Value, ok bool) {
	value, ok = c.vars[name]
	return
}

type ErrName struct {
	Name string
}

func (e *ErrName) Error() string {
	return fmt.Sprintf("name '%s' is not defined", e.Name)
}

func (c *Context) Lookup(name string) (value Value, err error) {
	value, ok := c.TryLookup(name)
	if !ok {
		if c.NameError != nil {
			value, err = c.NameError(name)
		} else {
			value, err = StringValue(""), nil
		}
	}
	return
}

func (c *Context) Declare(name string, value Value) {
	v := c.vars[name]
	c.shadowed = append(c.shadowed, namedVar{name, v})
	c.vars[name] = value
}

func (c *Context) Assign(name string, value Value) {
	c.vars[name] = value
}

func (c *Context) BeginScope() {
	c.prevScopes = append(c.prevScopes, c.scope)
	c.scope = len(c.shadowed)
}

func (c *Context) EndScope() {
	for i := len(c.shadowed) - 1; i >= c.scope; i-- {
		prevVar := c.shadowed[i]

		if prevVar.value == nil {
			delete(c.vars, prevVar.name)
		} else {
			c.vars[prevVar.name] = prevVar.value
		}
	}

	c.shadowed = c.shadowed[:c.scope]
	c.scope = c.prevScopes[len(c.prevScopes)-1]
	c.prevScopes = c.prevScopes[:len(c.prevScopes)-1]
}

type ValueWriter interface {
	WriteValue(v Value) error
}

func EvalRaw(c *Context, code []Instr, wr ValueWriter) (err error) {
	openScopes := 0
	defer func() {
		for openScopes > 0 {
			c.EndScope()
			openScopes--
		}
	}()

	pushedOutputFilters := 0
	defer func() {
		c.outputFilters = c.outputFilters[:len(c.outputFilters)-pushedOutputFilters]
	}()

	ip := 0
	var (
		stack valueStack
		value Value
	)

	for ip < len(code) {
		instr := code[ip]
		ip += 1

		switch instr.op {
		case emit:
			err = wr.WriteValue(StringValue(instr.sarg))
			if err != nil {
				return
			}
		case push:
			stack.Push(StringValue(instr.sarg))
		case emitFetch:
			value, err = c.Lookup(instr.sarg)
			if err != nil {
				return
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushFetch:
			value, err = c.Lookup(instr.sarg)
			if err != nil {
				return
			}
			stack.Push(value)
		case emitCall:
			err = evalCall(c, &stack, instr, wr)
			if err != nil {
				return err
			}
		case pushCall:
			retBuilder := returnValueBuilder{}
			err = evalCall(c, &stack, instr, &retBuilder)
			if err != nil {
				return err
			}
			stack.Push(retBuilder.Value())
		case emitCallDyn:
			err = evalCallDyn(c, &stack, instr, wr)
			if err != nil {
				return err
			}
		case pushCallDyn:
			retBuilder := returnValueBuilder{}
			err = evalCallDyn(c, &stack, instr, &retBuilder)
			if err != nil {
				return err
			}
			stack.Push(retBuilder.Value())
		case emitAttr:
			value, err = evalAttr(c, &stack, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushAttr:
			value, err = evalAttr(c, &stack, instr)
			if err != nil {
				return err
			}
			stack.Push(value)
		case emitSubprog:
			value, err = evalSubprog(c, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushSubprog:
			value, err = evalSubprog(c, instr)
			if err != nil {
				return err
			}
			stack.Push(value)
		case emitCompare:
			value, err = evalCompare(c, &stack, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushCompare:
			value, err = evalCompare(c, &stack, instr)
			if err != nil {
				return err
			}
			stack.Push(value)
		case jump:
			ip += instr.iarg
		case jumpTrue:
			if stack.Peek().Bool() {
				ip += instr.iarg
			}
		case jumpFalse:
			if !stack.Peek().Bool() {
				ip += instr.iarg
			}
		case emitPop:
			err = wr.WriteValue(stack.Pop())
			if err != nil {
				return err
			}
		case discardPop:
			stack.Pop()
		case storePop:
			value := stack.Pop()
			c.Assign(instr.sarg, value)
		case declarePop:
			value := stack.Pop()
			c.Declare(instr.sarg, value)
		case pushPeek:
			stack.Push(stack.Peek())
		case pushNot:
			pvalue := stack.PeekPtr()
			*pvalue = BoolValue(!(*pvalue).Bool())
		case emitNot:
			value = BoolValue(!stack.Pop().Bool())
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case emitBinaryOP:
			value, err = evalBinaryOP(c, &stack, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushBinaryOP:
			value, err = evalBinaryOP(c, &stack, instr)
			if err != nil {
				return err
			}
			stack.Push(value)
		case emitNumber:
			value = evalNumber(c, instr)
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushNumber:
			value = evalNumber(c, instr)
			stack.Push(value)
		case pushIter:
			value := stack.Pop()
			iter, err := value.Iter()
			if err != nil {
				return err
			}
			c.iters = append(c.iters, iter)
		case iterNextOrJump:
			ok := false
			value, ok, err = c.iters[len(c.iters)-1].Next()
			if err != nil {
				return
			}
			if ok {
				c.Assign(instr.sarg, value)
			} else {
				ip += instr.iarg
			}
		case discardIter:
			c.iters = c.iters[:len(c.iters)-1]
		case beginScope:
			c.BeginScope()
			openScopes++
		case endScope:
			c.EndScope()
			openScopes--
		case pushOutputFilter:
			pushedOutputFilters++
			c.outputFilters = append(c.outputFilters, c.valueFilters[instr.iarg])
		case popOutputFilter:
			pushedOutputFilters--
			c.outputFilters = c.outputFilters[:len(c.outputFilters)-1]

		}
	}
	return nil
}

func evalCall(c *Context, stack *valueStack, instr Instr, wr ValueWriter) (err error) {
	args := stack.PopN(instr.iarg)

	value, err := c.Lookup(instr.sarg)
	if err != nil {
		return
	}
	return value.Call(Args{args}, wr)
}

func evalCallDyn(c *Context, stack *valueStack, instr Instr, wr ValueWriter) (err error) {
	allArgs := stack.PopN(instr.iarg + 1)
	return allArgs[0].Call(Args{allArgs[1:]}, wr)
}

func evalAttr(c *Context, stack *valueStack, instr Instr) (value Value, err error) {
	value = stack.Pop()
	value, ok := value.GetAttr(instr.sarg)
	if !ok {
		value, err = c.NameError(instr.sarg)
	}
	return
}

func evalSubprog(c *Context, instr Instr) (value Value, err error) {
	subprog := c.subprogs[instr.iarg]
	ctx := c.Clone()
	value = &subprogValue{subprog.Code, subprog.Args, ctx}
	return
}

func evalCompare(c *Context, stack *valueStack, instr Instr) (value Value, err error) {
	args := stack.PopN(2)

	ok, err := compareValues(args[0], args[1], instr.iarg)
	value = BoolValue(ok)
	return
}

func evalBinaryOP(c *Context, stack *valueStack, instr Instr) (value Value, err error) {
	args := stack.PopN(2)

	value, err = binaryOPValues(args[0], args[1], instr.iarg)
	return
}

func evalNumber(c *Context, instr Instr) (value Value) {
	number, _ := strconv.ParseFloat(instr.sarg, 64) // error was tested during parsing
	value = NumberValue(number)
	return
}

type stringBuilder struct {
	c *Context
	b strings.Builder
}

func (b *stringBuilder) WriteValue(v Value) error {
	str, err := v.String()
	if err != nil {
		return err
	}
	if len(b.c.outputFilters) > 0 {
		if f := b.c.outputFilters[len(b.c.outputFilters)-1]; f != nil {
			str, err = f.Filter(str)
			if err != nil {
				return err
			}
		}
	}
	b.b.WriteString(str)
	return nil
}

func (b *stringBuilder) String() string {
	return b.b.String()
}

func EvalString(c *Context, code []Instr) (string, error) {
	b := stringBuilder{c: c}
	err := EvalRaw(c, code, &b)
	return b.String(), err
}

type outputWriter struct {
	c *Context
	w io.Writer
}

func (w *outputWriter) WriteValue(v Value) error {
	str, err := v.String()
	if err != nil {
		return err
	}
	if len(w.c.outputFilters) > 0 {
		if f := w.c.outputFilters[len(w.c.outputFilters)-1]; f != nil {
			str, err = f.Filter(str)
			if err != nil {
				return err
			}
		}
	}
	n, err := w.w.Write([]byte(str))
	if err == nil && n < len(str) {
		err = io.ErrShortWrite
	}
	return err
}

func EvalWriter(c *Context, code []Instr) error {
	w := outputWriter{}
	err := EvalRaw(c, code, &w)
	return err
}

type returnValueBuilder struct {
	hasValue bool
	value    Value
	sb       strings.Builder
}

func (w *returnValueBuilder) WriteValue(v Value) error {
	if !w.hasValue {
		w.value = v
		w.hasValue = true
		return nil
	}

	if w.value != nil {
		str, err := w.value.String()
		w.value = nil
		if err != nil {
			return err
		}
		w.sb.WriteString(str)
	}

	str, err := v.String()
	if err != nil {
		return err
	}
	w.sb.WriteString(str)
	return nil
}

func (w *returnValueBuilder) Value() Value {
	if !w.hasValue {
		return StringValue("")
	}
	if w.value != nil {
		return w.value
	}
	return StringValue(w.sb.String())
}

func Call(v Value, args []Value) (Value, error) {
	wr := returnValueBuilder{}
	err := v.Call(Args{args}, &wr)
	return wr.Value(), err
}

type valueStack struct {
	stack []Value
}

func (s *valueStack) Push(v Value) {
	s.stack = append(s.stack, v)
}

func (s *valueStack) Peek() Value {
	return s.stack[len(s.stack)-1]
}

func (s *valueStack) PeekPtr() *Value {
	return &s.stack[len(s.stack)-1]
}

func (s *valueStack) Pop() Value {
	value := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return value
}

func (s *valueStack) PopN(n int) []Value {
	values := s.stack[len(s.stack)-n:]
	s.stack = s.stack[:len(s.stack)-n]
	return values
}

type singleValueIter struct {
	v Value
}

var _ ValueIter = &singleValueIter{}

func (i *singleValueIter) Next() (Value, bool, error) {
	if i.v != nil {
		v := i.v
		i.v = nil
		return v, true, nil
	}
	return nil, false, nil
}

type listIter struct {
	list []Value
}

var _ ValueIter = &listIter{}

func (l *listIter) Next() (Value, bool, error) {
	if len(l.list) > 0 {
		v := l.list[0]
		l.list = l.list[1:]
		return v, true, nil
	}
	return nil, false, nil
}
