package tplexpr

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Reflect(v interface{}) Value {
	switch v := v.(type) {
	case nil:
		return Nil
	case bool:
		return BoolValue(v)
	case int:
		return NumberValue(v)
	case int8:
		return NumberValue(v)
	case int16:
		return NumberValue(v)
	case int32:
		return NumberValue(v)
	case int64:
		return NumberValue(v)
	case uint:
		return NumberValue(v)
	case uint8:
		return NumberValue(v)
	case uint16:
		return NumberValue(v)
	case uint32:
		return NumberValue(v)
	case uint64:
		return NumberValue(v)
	case float32:
		return NumberValue(v)
	case float64:
		return NumberValue(v)
	case string:
		return StringValue(v)
	case []Value:
		return ListValue(v)
	case map[string]Value:
		return ObjectValue(v)
	}

	rv := reflect.ValueOf(v)
	for {
		switch rv.Kind() {
		case reflect.Bool:
			return reflectBool{rv}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflectNumber{rv: rv, number: float64(rv.Int())}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return reflectNumber{rv: rv, number: float64(rv.Uint())}
		case reflect.Float32, reflect.Float64:
			return reflectNumber{rv: rv, number: rv.Float()}
		case reflect.Array, reflect.Slice:
			return reflectList{rv: rv}
		case reflect.Chan:
			return reflectChan{rv: rv}
		case reflect.Func, reflect.Interface, reflect.Map, reflect.Struct:
			return &reflectObjectValue{obj: reflectObject{rv: rv}}
		case reflect.Pointer:
			if rv.IsNil() {
				return Nil
			}
			rv = rv.Elem()
		case reflect.String:
			return reflectString{rv: rv}
		default:
			return Nil
		}
	}
}

type reflectBool struct {
	rv reflect.Value
}

var _ Value = reflectBool{}

func (v reflectBool) Kind() ValueKind {
	return KindBool
}

func (v reflectBool) Bool() bool {
	return v.rv.Bool()
}

func (v reflectBool) Number() (float64, error) {
	if v.Bool() {
		return 1, nil
	}
	return 0, nil
}

func (v reflectBool) String() (string, error) {
	if s, ok := v.rv.Interface().(fmt.Stringer); ok {
		return s.String(), nil
	}
	return fmt.Sprintf("%v", v.rv.Bool()), nil
}

func (v reflectBool) List() ([]Value, error) {
	return []Value{v}, nil
}

func (v reflectBool) Iter() (ValueIter, error) {
	return &listIter{[]Value{v}}, nil
}

func (v reflectBool) Object() (Object, error) {
	return &reflectObject{rv: v.rv}, nil
}

func (v reflectBool) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(v)
}

type reflectNumber struct {
	number float64
	rv     reflect.Value
}

func (v reflectNumber) Kind() ValueKind {
	return KindNumber
}

func (v reflectNumber) Bool() bool {
	return v.number != 0
}

func (v reflectNumber) Number() (float64, error) {
	return v.number, nil
}

func (v reflectNumber) String() (string, error) {
	if s, ok := v.rv.Interface().(fmt.Stringer); ok {
		return s.String(), nil
	}

	return fmt.Sprintf("%v", v.number), nil
}

func (v reflectNumber) List() ([]Value, error) {
	return []Value{v}, nil
}

func (v reflectNumber) Iter() (ValueIter, error) {
	return &listIter{[]Value{v}}, nil
}

func (v reflectNumber) Object() (Object, error) {
	return &reflectObject{rv: v.rv}, nil
}

func (v reflectNumber) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(v)
}

type reflectList struct {
	rv reflect.Value
}

func (v reflectList) toList() []Value {
	l := v.rv.Len()
	lst := make([]Value, l)
	for i := 0; i < l; i++ {
		lst[i] = Reflect(v.rv.Index(i).Interface())
	}
	return lst
}

func (v reflectList) Kind() ValueKind {
	return KindList
}

func (v reflectList) Bool() bool {
	return v.rv.Len() != 0
}

func (v reflectList) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindListName, conTO, KindNumberName}
}

func (v reflectList) String() (string, error) {
	if s, ok := v.rv.Interface().(fmt.Stringer); ok {
		return s.String(), nil
	}
	lst := v.toList()
	slst := make([]string, len(lst))
	var err error
	for i := range lst {
		slst[i], err = lst[i].String()
		if err != nil {
			return "", err
		}
	}
	return strings.Join(slst, " "), nil
}

func (v reflectList) List() ([]Value, error) {
	return v.toList(), nil
}

func (v reflectList) Iter() (ValueIter, error) {
	return &listIter{v.toList()}, nil
}

func (v reflectList) Object() (Object, error) {
	return &reflectObject{rv: v.rv}, nil
}

func (v reflectList) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(v)
}

type reflectObject struct {
	m  map[string]Value
	rv reflect.Value
}

var _ Object = &reflectObject{}

func (o *reflectObject) Key(name string) (Value, bool) {
	if v, ok := o.m[name]; ok {
		return v, true
	}

	switch o.rv.Kind() {
	case reflect.Array, reflect.Slice:
		idx, err := strconv.ParseInt(name, 10, 64)
		if err == nil && idx >= 0 && idx < int64(o.rv.Len()) {
			return Reflect(o.rv.Index(int(idx))), true
		}
	case reflect.Map:
		typ := o.rv.Type()
		if reflect.TypeOf("").ConvertibleTo(typ.Key()) {
			key := reflect.ValueOf(name).Convert(typ.Key())
			v := o.rv.MapIndex(key)
			if v != (reflect.Value{}) {
				return Reflect(v.Interface()), true
			}
		}
	case reflect.Struct:
		typ := o.rv.Type()
		field, ok := typ.FieldByName(name)
		if !ok || !field.IsExported() {
			break
		}
		return Reflect(o.rv.FieldByName(field.Name).Interface()), true
	}

	typ := o.rv.Type()
	method, ok := typ.MethodByName(name)
	if !ok {
		return nil, false
	}
	if method.IsExported() && method.Type.NumIn() == 0 && method.Type.NumOut() == 1 {
		v := o.rv.MethodByName(name).Call(nil)[0]
		return Reflect(v.Interface()), true
	}
	return nil, false
}

func (o *reflectObject) Keys() []string {
	keys := map[string]struct{}{}

	for key := range o.m {
		keys[key] = struct{}{}
	}

	switch o.rv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := o.rv.Len() - 1; i >= 0; i-- {
			keys[fmt.Sprintf("%d", i)] = struct{}{}
		}

	case reflect.Map:
		typ := o.rv.Type()
		if reflect.TypeOf("").ConvertibleTo(typ.Key()) {
			for _, key := range o.rv.MapKeys() {
				keys[fmt.Sprintf("%v", key.Interface())] = struct{}{}
			}
		}
	case reflect.Struct:
		typ := o.rv.Type()
		for i := typ.NumField() - 1; i >= 0; i-- {
			field := typ.Field(i)
			if field.IsExported() {
				keys[field.Name] = struct{}{}
			}
		}
	}

	// get methods
	typ := o.rv.Type()
	for i := typ.NumMethod() - 1; i >= 0; i-- {
		method := typ.Method(i)
		if method.IsExported() && method.Type.NumIn() == 0 && method.Type.NumOut() == 1 {
			keys[method.Name] = struct{}{}
		}
	}

	keyList := make([]string, 0, len(keys))
	for key := range keys {
		keyList = append(keyList, key)
	}
	return keyList
}

func (o *reflectObject) SetKey(name string, value Value) {
	if o.m == nil {
		o.m = map[string]Value{}
	}
	o.m[name] = value
}

type reflectChanIter struct {
	ch reflect.Value
}

var _ ValueIter = reflectChanIter{}

func (i reflectChanIter) Next() (v Value, ok bool, err error) {
	rcv, ok := i.ch.Recv()
	return Reflect(rcv), ok, nil
}

type reflectChan struct {
	rv reflect.Value
}

func (v reflectChan) Kind() ValueKind {
	return KindIterator
}

func (v reflectChan) Bool() bool {
	return true
}

func (v reflectChan) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindIteratorName, conTO, KindNumberName}
}

func (v reflectChan) String() (string, error) {
	if s, ok := v.rv.Interface().(fmt.Stringer); ok {
		return s.String(), nil
	}
	return "", nil
}

func (v reflectChan) List() ([]Value, error) {
	lst := []Value{}
	for {
		v, ok := v.rv.Recv()
		if !ok {
			break
		}
		lst = append(lst, Reflect(v))
	}
	return lst, nil
}

func (v reflectChan) Iter() (ValueIter, error) {
	return reflectChanIter{v.rv}, nil
}

func (v reflectChan) Object() (Object, error) {
	return &reflectObject{rv: v.rv}, nil
}

func (v reflectChan) Call(args Args, wr ValueWriter) error {
	r, ok := v.rv.Recv()
	if ok {
		return wr.WriteValue(Reflect(r.Interface()))
	}
	return nil
}

type reflectObjectValue struct {
	obj reflectObject
}

func (v *reflectObjectValue) toList() []Value {
	keys := v.obj.Keys()
	lst := make([]Value, len(keys))
	for i := range keys {
		lst[i] = StringValue(keys[i])
	}
	return lst
}

func (v *reflectObjectValue) Kind() ValueKind {
	return KindObject
}

func (v *reflectObjectValue) Bool() bool {
	return true
}

func (v *reflectObjectValue) Number() (float64, error) {
	return 0, &ErrType{opConvert, KindObjectName, conTO, KindNumberName}
}

func (v *reflectObjectValue) String() (string, error) {
	if s, ok := v.obj.rv.Interface().(fmt.Stringer); ok {
		return s.String(), nil
	}

	keys := v.obj.Keys()
	return strings.Join(keys, " "), nil
}

func (v *reflectObjectValue) List() ([]Value, error) {
	return v.toList(), nil
}

func (v *reflectObjectValue) Iter() (ValueIter, error) {
	return &listIter{v.toList()}, nil
}

func (v *reflectObjectValue) Object() (Object, error) {
	return &v.obj, nil
}

func (v *reflectObjectValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(v)
}

type reflectString struct {
	rv reflect.Value
}

func (v reflectString) toString() string {
	if s, ok := v.rv.Interface().(fmt.Stringer); ok {
		return s.String()
	}
	return v.rv.String()
}

func (v reflectString) Kind() ValueKind {
	return KindString
}

func (v reflectString) Bool() bool {
	return len(v.toString()) != 0
}

func (v reflectString) Number() (float64, error) {
	return strconv.ParseFloat(v.toString(), 64)
}

func (v reflectString) String() (string, error) {
	return v.toString(), nil
}

func (v reflectString) List() ([]Value, error) {
	return []Value{v}, nil
}

func (v reflectString) Iter() (ValueIter, error) {
	return &listIter{[]Value{v}}, nil
}

func (v reflectString) Object() (Object, error) {
	return &reflectObject{rv: v.rv}, nil
}

func (v reflectString) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(v)
}
