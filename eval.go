package tplexpr

import (
	"fmt"
	"strconv"
	"strings"
)

type ValueKind int

const (
	KindString ValueKind = iota
	KindList
	KindObject
	KindFunction
)

type Value interface {
	Kind() ValueKind
	String() (string, error)
	List() ([]Value, error)
	GetAttr(name string) (Value, bool)
	Call(args []Value) (Value, error)
}

type StringValue string

var _ Value = StringValue("")

func (s StringValue) Kind() ValueKind {
	return KindString
}

func (s StringValue) String() (string, error) {
	return string(s), nil
}

func (s StringValue) List() ([]Value, error) {
	return []Value{s}, nil
}

func (s StringValue) GetAttr(name string) (Value, bool) {
	return nil, false
}

func (s StringValue) Call(args []Value) (Value, error) {
	return s, nil
}

type ListValue []Value

var _ Value = ListValue{}

func (l ListValue) Kind() ValueKind {
	return KindList
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

func (l ListValue) Call(args []Value) (Value, error) {
	return l, nil
}

type ObjectValue map[string]Value

var _ Value = ObjectValue{}

func (o ObjectValue) Kind() ValueKind {
	return KindObject
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

func (o ObjectValue) GetAttr(name string) (v Value, ok bool) {
	v, ok = o[name]
	return
}

func (o ObjectValue) Call(args []Value) (Value, error) {
	return o, nil
}

type FuncValue func(args []Value) (Value, error)

var _ Value = FuncValue(nil)

func (f FuncValue) Kind() ValueKind {
	return KindFunction
}

func (f FuncValue) String() (string, error) {
	value, err := f(nil)
	if err != nil {
		return "", err
	}
	return value.String()
}

func (f FuncValue) List() ([]Value, error) {
	value, err := f(nil)
	if err != nil {
		return nil, err
	}
	return value.List()
}

func (f FuncValue) GetAttr(name string) (Value, bool) {
	return nil, false
}

func (f FuncValue) Call(args []Value) (Value, error) {
	return f(args)
}

type subprogValue struct {
	code []Instr
	args []string
	ctx  *Context
}

var _ Value = &subprogValue{}

func (s *subprogValue) eval(args []Value) (Value, error) {
	s.ctx.BeginScope()
	defer s.ctx.EndScope()

	for i, argName := range s.args {
		s.ctx.Declare(argName, GetArg(args, i, StringValue("")))
	}

	str, err := EvalString(s.ctx, s.code)
	return StringValue(str), err
}

func (s *subprogValue) Kind() ValueKind {
	return KindFunction
}

func (s *subprogValue) String() (string, error) {
	value, err := s.eval(nil)
	if err != nil {
		return "", err
	}
	return value.String()
}

func (s *subprogValue) List() ([]Value, error) {
	value, err := s.eval(nil)
	if err != nil {
		return nil, err
	}
	return value.List()
}

func (s *subprogValue) GetAttr(name string) (Value, bool) {
	return nil, false
}

func (s *subprogValue) Call(args []Value) (Value, error) {
	return s.eval(args)
}

type Subprog struct {
	Args []string
	Code []Instr
}

type Context struct {
	vars       map[string]Value
	shadowed   []namedVar
	scope      int
	prevScopes []int
	subprogs   []Subprog
	NameError  func(name string) (Value, error)
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

func Eval(c *Context, code []Instr, wr ValueWriter) (err error) {
	ip := 0
	var (
		stack []Value
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
			stack = append(stack, StringValue(instr.sarg))
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
			stack = append(stack, value)
		case emitCall:
			value, err = evalCall(c, &stack, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushCall:
			value, err = evalCall(c, &stack, instr)
			if err != nil {
				return err
			}
			stack = append(stack, value)
		case emitCallDyn:
			value, err = evalCallDyn(c, &stack, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushCallDyn:
			value, err = evalCallDyn(c, &stack, instr)
			if err != nil {
				return err
			}
			stack = append(stack, value)

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
			stack = append(stack, value)
		case emitSubprog:
			value, err = evalSubprog(c, &stack, instr)
			if err != nil {
				return err
			}
			err = wr.WriteValue(value)
			if err != nil {
				return err
			}
		case pushSubprog:
			value, err = evalSubprog(c, &stack, instr)
			if err != nil {
				return err
			}
			stack = append(stack, value)
		}
	}
	return nil
}

func evalCall(c *Context, stack *[]Value, instr Instr) (value Value, err error) {
	args := (*stack)[len(*stack)-instr.iarg:]
	*stack = (*stack)[:len(args)]

	value, err = c.Lookup(instr.sarg)
	if err != nil {
		return
	}
	value, err = value.Call(args)
	return
}

func evalCallDyn(c *Context, stack *[]Value, instr Instr) (value Value, err error) {
	allArgs := (*stack)[len(*stack)-instr.iarg-1:]
	*stack = (*stack)[:len(*stack)-instr.iarg-1]

	value, err = allArgs[0].Call(allArgs[1:])
	return
}

func evalAttr(c *Context, stack *[]Value, instr Instr) (value Value, err error) {
	value = (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	value, ok := value.GetAttr(instr.sarg)
	if !ok {
		value, err = c.NameError(instr.sarg)
	}
	return
}

func evalSubprog(c *Context, stack *[]Value, instr Instr) (value Value, err error) {
	subprog := c.subprogs[instr.iarg]
	ctx := c.Clone()
	value = &subprogValue{subprog.Code, subprog.Args, ctx}
	return
}

type stringBuilder struct {
	b strings.Builder
}

func (b *stringBuilder) WriteValue(v Value) error {
	str, err := v.String()
	if err == nil {
		b.b.WriteString(str)
	}
	return err
}

func (b *stringBuilder) String() string {
	return b.b.String()
}

func EvalString(c *Context, code []Instr) (string, error) {
	b := stringBuilder{}
	err := Eval(c, code, &b)
	return b.String(), err
}
