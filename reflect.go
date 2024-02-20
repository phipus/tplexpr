package tplexpr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type ReflectValue struct {
	v interface{}
}

var _ Value = &ReflectValue{}

func Reflect(v interface{}) (*ReflectValue, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	value := &ReflectValue{}
	err = json.Unmarshal(data, &value.v)
	return value, err
}

func (v *ReflectValue) Kind() ValueKind {
	switch v.v.(type) {
	case nil:
		return KindString
	case bool:
		return KindBool
	case float64:
		return KindNumber
	case string:
		return KindString
	case []interface{}:
		return KindList
	case map[string]interface{}:
		return KindObject
	default:
		return KindString
	}
}

func (v *ReflectValue) Bool() bool {
	switch v := v.v.(type) {
	case nil:
		return false
	case bool:
		return v
	case float64:
		return v != 0
	case string:
		return len(v) != 0
	case []interface{}:
		return len(v) != 0
	case map[string]interface{}:
		return true
	default:
		return len(fmt.Sprintf("%v", v)) != 0
	}
}

func (v *ReflectValue) Number() (float64, error) {
	switch v := v.v.(type) {
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	case []interface{}:
		return 0, &ErrType{opConvert, KindListName, conTO, KindNumberName}
	case map[string]interface{}:
		return 0, &ErrType{opConvert, KindObjectName, conTO, KindNumberName}
	case nil:
		return strconv.ParseFloat("", 64)
	default:
		return strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	}
}

func (v *ReflectValue) String() (string, error) {
	return reflectValueString(v.v)
}

func reflectValueString(v interface{}) (string, error) {
	switch v := v.(type) {
	case nil:
		return "", nil
	case float64:
		return fmt.Sprintf("%g", v), nil
	case string:
		return v, nil
	case []interface{}:
		var err error
		keys := make([]string, len(v))
		for i := range v {
			keys[i], err = reflectValueString(v[i])
			if err != nil {
				return "", err
			}
		}
		return strings.Join(keys, " "), nil
	case map[string]interface{}:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		return strings.Join(keys, " "), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (v *ReflectValue) List() ([]Value, error) {
	switch v := v.v.(type) {
	case nil:
		return []Value{}, nil
	case bool:
		return []Value{BoolValue(v)}, nil
	case float64:
		return []Value{NumberValue(v)}, nil
	case string:
		return []Value{StringValue(v)}, nil
	case []interface{}:
		lst := make([]Value, len(v))
		for i := range v {
			lst[i] = &ReflectValue{v[i]}
		}
		return lst, nil
	case map[string]interface{}:
		keys := make([]Value, 0, len(v))
		for key := range v {
			keys = append(keys, StringValue(key))
		}
		return keys, nil
	default:
		return []Value{StringValue(fmt.Sprintf("%v", v))}, nil
	}
}

func (v *ReflectValue) Iter() (ValueIter, error) {
	lst, err := v.List()
	return &listIter{lst}, err
}

func (v *ReflectValue) Object() (Object, error) {
	switch v := v.v.(type) {
	case []interface{}:
		lst := make([]Value, len(v))
		for i := range v {
			lst[i] = &ReflectValue{v[i]}
		}
		return &ListObject{lst}, nil
	case map[string]interface{}:
		obj := &MapObject{M: map[string]Value{}}
		for key, value := range v {
			obj.M[key] = &ReflectValue{value}
		}
		return obj, nil
	default:
		return &MapObject{}, nil
	}
}

func (v *ReflectValue) Call(args Args, wr ValueWriter) error {
	return wr.WriteValue(v)
}
