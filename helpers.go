package tplexpr

type (
	B = BoolValue
	N = NumberValue
	S = StringValue
	L = ListValue
	O = ObjectValue
	F = FuncValue
)

type Vars map[string]Value

func BuildVars() Vars {
	return nil
}

func (v Vars) Set(name string, value Value) Vars {
	if v == nil {
		v = Vars{}
	}
	v[name] = value
	return v
}

func (v Vars) SetString(name, value string) Vars {
	return v.Set(name, StringValue(value))
}

func (v Vars) SetBool(name string, value bool) Vars {
	return v.Set(name, BoolValue(value))
}

func (v Vars) SetNumber(name string, value float64) Vars {
	return v.Set(name, NumberValue(value))
}

func BuildObject() ObjectValue {
	return nil
}

func (o ObjectValue) Set(name string, value Value) ObjectValue {
	if o == nil {
		o = ObjectValue{}
	}
	o[name] = value
	return o
}

func (o ObjectValue) SetString(name, value string) ObjectValue {
	return o.Set(name, StringValue(value))
}

func (o ObjectValue) SetBool(name string, value bool) ObjectValue {
	return o.Set(name, BoolValue(value))
}

func (o ObjectValue) SetNumber(name string, value float64) ObjectValue {
	return o.Set(name, NumberValue(value))
}

func BuildList(values ...Value) ListValue {
	return values
}

func (l ListValue) Add(value Value) ListValue {
	return append(l, value)
}

func (l ListValue) AddString(value string) ListValue {
	return l.Add(StringValue(value))
}

func (l ListValue) AddBool(value bool) ListValue {
	return l.Add(BoolValue(value))
}

func (l ListValue) AddNumber(value float64) ListValue {
	return l.Add(NumberValue(value))
}

func (l ListValue) AddStrings(value []string) ListValue {
	for _, v := range value {
		l = append(l, StringValue(v))
	}
	return l
}
