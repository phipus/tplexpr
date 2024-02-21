package tplexpr

type (
	B = BoolValue
	N = NumberValue
	S = StringValue
	L = ListValue
	O = ObjectValue
	F = FuncValue
)

type VarsBuilder struct {
	vars Vars
	err  error
}

type Vars map[string]Value

func BuildVars() *VarsBuilder {
	return &VarsBuilder{vars: Vars{}}
}

func (b *VarsBuilder) Set(name string, value Value) *VarsBuilder {
	b.vars[name] = value
	return b
}

func (b *VarsBuilder) SetString(name, value string) *VarsBuilder {
	return b.Set(name, StringValue(value))
}

func (b *VarsBuilder) SetBool(name string, value bool) *VarsBuilder {
	return b.Set(name, BoolValue(value))
}

func (b *VarsBuilder) SetNumber(name string, value float64) *VarsBuilder {
	return b.Set(name, NumberValue(value))
}

func (b *VarsBuilder) SetList(name string, value *ListBuilder) *VarsBuilder {
	if b.err != nil {
		return b
	}
	var v ListValue
	v, b.err = value.Build()
	return b.Set(name, v)
}

func (b *VarsBuilder) SetObject(name string, value *ObjectBuilder) *VarsBuilder {
	if b.err != nil {
		return b
	}
	var v ObjectValue
	v, b.err = value.Build()
	return b.Set(name, v)
}

func (b *VarsBuilder) SetReflect(name string, value interface{}) *VarsBuilder {
	if b.err != nil {
		return b
	}
	var v Value
	v, b.err = Reflect(value)
	return b.Set(name, v)
}

func (b *VarsBuilder) SetMap(m map[string]Value) *VarsBuilder {
	for name, value := range m {
		b.vars[name] = value
	}
	return b
}

func (b *VarsBuilder) Build() (Vars, error) {
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
