package tplexpr

type (
	B = BoolValue
	N = NumberValue
	S = StringValue
	L = ListValue
	O = ObjectValue
	F = FuncValue
)

type ScopeBuilder struct {
	vars VarScope
	err  error
}

type VarScope map[string]Value

func BuildScope() *ScopeBuilder {
	return &ScopeBuilder{vars: VarScope{}}
}

func (b *ScopeBuilder) Set(name string, value Value) *ScopeBuilder {
	b.vars[name] = value
	return b
}

func (b *ScopeBuilder) SetString(name, value string) *ScopeBuilder {
	return b.Set(name, StringValue(value))
}

func (b *ScopeBuilder) SetBool(name string, value bool) *ScopeBuilder {
	return b.Set(name, BoolValue(value))
}

func (b *ScopeBuilder) SetNumber(name string, value float64) *ScopeBuilder {
	return b.Set(name, NumberValue(value))
}

func (b *ScopeBuilder) SetList(name string, value *ListBuilder) *ScopeBuilder {
	if b.err != nil {
		return b
	}
	var v ListValue
	v, b.err = value.Build()
	return b.Set(name, v)
}

func (b *ScopeBuilder) SetObject(name string, value *ObjectBuilder) *ScopeBuilder {
	if b.err != nil {
		return b
	}
	var v ObjectValue
	v, b.err = value.Build()
	return b.Set(name, v)
}

func (b *ScopeBuilder) SetReflect(name string, value interface{}) *ScopeBuilder {
	if b.err != nil {
		return b
	}
	var v Value
	v, b.err = Reflect(value)
	return b.Set(name, v)
}

func (b *ScopeBuilder) SetMap(m map[string]Value) *ScopeBuilder {
	for name, value := range m {
		b.vars[name] = value
	}
	return b
}

func (b *ScopeBuilder) Build() (VarScope, error) {
	return b.vars, b.err
}

type ObjectBuilder struct {
	obj ObjectValue
	err error
}

func BuildObject() *ObjectBuilder {
	return &ObjectBuilder{obj: ObjectValue{}}
}

func (b *ObjectBuilder) Set(name string, value Value) *ObjectBuilder {
	b.obj[name] = value
	return b
}

func (b *ObjectBuilder) SetString(name, value string) *ObjectBuilder {
	return b.Set(name, StringValue(value))
}

func (b *ObjectBuilder) SetBool(name string, value bool) *ObjectBuilder {
	return b.Set(name, BoolValue(value))
}

func (b *ObjectBuilder) SetNumber(name string, value float64) *ObjectBuilder {
	return b.Set(name, NumberValue(value))
}

func (b *ObjectBuilder) SetList(name string, value *ListBuilder) *ObjectBuilder {
	if b.err != nil {
		return b
	}
	var v ListValue
	v, b.err = value.Build()
	return b.Set(name, v)
}

func (b *ObjectBuilder) SetObject(name string, value *ObjectBuilder) *ObjectBuilder {
	if b.err != nil {
		return b
	}
	var v ObjectValue
	v, b.err = value.Build()
	return b.Set(name, v)
}

func (b *ObjectBuilder) SetReflect(name string, value interface{}) *ObjectBuilder {
	if b.err != nil {
		return b
	}
	var v Value
	v, b.err = Reflect(value)
	return b.Set(name, v)
}

func (b *ObjectBuilder) SetMap(m map[string]Value) *ObjectBuilder {
	for name, value := range m {
		b.obj[name] = value
	}
	return b
}

func (b *ObjectBuilder) Build() (ObjectValue, error) {
	return b.obj, b.err
}

type ListBuilder struct {
	lst ListValue
	err error
}

func BuildList(values ...Value) *ListBuilder {
	return &ListBuilder{lst: values}
}

func (b *ListBuilder) Add(value Value) *ListBuilder {
	b.lst = append(b.lst, value)
	return b
}

func (b *ListBuilder) AddString(value string) *ListBuilder {
	return b.Add(StringValue(value))
}

func (b *ListBuilder) AddBool(value bool) *ListBuilder {
	return b.Add(BoolValue(value))
}

func (b *ListBuilder) AddNumber(value float64) *ListBuilder {
	return b.Add(NumberValue(value))
}

func (b *ListBuilder) AddList(value *ListBuilder) *ListBuilder {
	if b.err != nil {
		return b
	}
	var v ListValue
	v, b.err = value.Build()
	return b.Add(v)
}

func (b *ListBuilder) AddObject(value *ObjectBuilder) *ListBuilder {
	if b.err != nil {
		return b
	}
	var v ObjectValue
	v, b.err = value.Build()
	return b.Add(v)
}

func (b *ListBuilder) AddReflect(value interface{}) *ListBuilder {
	if b.err != nil {
		return b
	}
	var v Value
	v, b.err = Reflect(value)
	return b.Add(v)
}

func (l ListValue) AddStrings(value []string) ListValue {
	for _, v := range value {
		l = append(l, StringValue(v))
	}
	return l
}

func (b *ListBuilder) Extend(lst []Value) *ListBuilder {
	b.lst = append(b.lst, lst...)
	return b
}

func (b *ListBuilder) Build() (ListValue, error) {
	return b.lst, b.err
}
