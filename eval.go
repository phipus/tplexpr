package tplexpr

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Subprog struct {
	Args []string
	Code []Instr
}

type Context struct {
	vars             map[string]Value
	shadowed         []namedVar
	scope            int
	prevScopes       []int
	subprogs         []Subprog
	iters            []ValueIter
	valueFilters     []ValueFilter
	outputFilters    []ValueFilter
	templates        map[string]Template
	NameError        func(name string) (Value, error)
	TemplateNotFound func(name string) error
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
	clone.valueFilters = c.valueFilters
	clone.templates = c.templates
	clone.NameError = c.NameError
	clone.TemplateNotFound = c.TemplateNotFound
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
			value = Nil
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
		case emitCallSubprogNA:
			err = evalCallSubprogNA(c, &stack, instr, wr)
			if err != nil {
				return err
			}
		case pushCallSubprogNA:
			retBuilder := returnValueBuilder{}
			err = evalCallSubprogNA(c, &stack, instr, &retBuilder)
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
		case includeTemplate:
			err = evalTemplate(c, instr.sarg, wr)
			if err != nil {
				return
			}
		case includeTemplateDyn:
			nameValue := stack.Pop()
			name, err := nameValue.String()
			if err != nil {
				return err
			}
			err = evalTemplate(c, name, wr)
			if err != nil {
				return err
			}
		case assignKey:
			value := stack.Pop()

			if obj, ok := stack.Peek().(ObjectValue); ok && obj != nil {
				obj[instr.sarg] = value
			} else if om, ok := stack.Peek().(objectMapper); ok {
				om.o.SetKey(instr.sarg, value)
			} else {
				obj, err := stack.Pop().Object()
				if err != nil {
					return err
				}
				obj.SetKey(instr.sarg, value)
				stack.Push(objectMapper{obj})
			}
		case assignKeyDyn:
			value := stack.Pop()
			key := stack.Pop()
			keyStr, err := key.String()
			if err != nil {
				return err
			}

			if obj, ok := stack.Peek().(ObjectValue); ok && obj != nil {
				obj[keyStr] = value
			} else {
				obj, err := stack.Pop().Object()
				if err != nil {
					return err
				}

				obj.SetKey(keyStr, value)
				stack.Push(objectMapper{obj})
			}
		case pushObject:
			stack.Push(ObjectValue{})
		case extendObject:
			obj, err := stack.Pop().Object()
			if err != nil {
				return err
			}
			newObj := ObjectValue{}
			switch obj := obj.(type) {
			case *MapObject:
				for key, value := range obj.M {
					newObj[key] = value
				}
			case *ListObject:
				for i, v := range obj.L {
					newObj[fmt.Sprintf("%d", i)] = v
				}
			default:
				for _, key := range obj.Keys() {
					newObj[key], _ = obj.Key(key)
				}
			}
			stack.Push(newObj)
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

func evalCallSubprogNA(c *Context, stack *valueStack, instr Instr, wr ValueWriter) error {
	c.BeginScope()
	defer c.EndScope()

	return EvalRaw(c, c.subprogs[instr.iarg].Code, wr)
}

func evalAttr(c *Context, stack *valueStack, instr Instr) (value Value, err error) {
	value = stack.Pop()
	obj, err := value.Object()
	if err != nil {
		return nil, err
	}
	value, ok := obj.Key(instr.sarg)
	if !ok {
		if c.NameError != nil {
			value, err = c.NameError(instr.sarg)
		} else {
			value = Nil
		}
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

func evalTemplate(c *Context, name string, wr ValueWriter) (err error) {
	tpl, ok := c.templates[name]
	if ok {
		err = EvalRaw(c, tpl.Code, wr)
	} else if c.TemplateNotFound != nil {
		err = c.TemplateNotFound(name)
	}
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

func (c *Context) EvalTemplateRaw(name string, vars Vars, wr ValueWriter) error {
	c.BeginScope()
	defer c.EndScope()

	for name, value := range vars {
		c.Declare(name, value)
	}

	return evalTemplate(c, name, wr)
}

func (c *Context) EvalTemplateString(name string, vars Vars) (string, error) {
	b := stringBuilder{c: c}
	err := c.EvalTemplateRaw(name, vars, &b)
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

func EvalWriter(c *Context, code []Instr, wr io.Writer) error {
	w := outputWriter{c: c, w: wr}
	err := EvalRaw(c, code, &w)
	return err
}

func (c *Context) EvalTemplateWriter(name string, vars Vars, wr io.Writer) error {
	w := outputWriter{c: c, w: wr}
	err := c.EvalTemplateRaw(name, vars, &w)
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
		return Nil
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
