package tplexpr

func compareValues(a, b Value, cmp int) (ok bool, err error) {
	if a.Kind() != b.Kind() {
		switch cmp {
		case NE:
			return true, nil
		default:
			return false, nil
		}
	}

	switch a.Kind() {
	case KindString:
		l, err := a.String()
		if err != nil {
			return ok, err
		}
		r, err := b.String()
		if err != nil {
			return ok, err
		}
		switch cmp {
		case EQ:
			ok = l == r
		case NE:
			ok = l != r
		case GT:
			ok = l > r
		case GE:
			ok = l >= r
		case LT:
			ok = l < r
		case LE:
			ok = l <= r
		}
		return ok, nil
	case KindBool:
		switch cmp {
		case EQ:
			ok = a.Bool() == b.Bool()
		case NE:
			ok = a.Bool() != b.Bool()
		}
		return
	case KindNumber:
		l, err := a.Number()
		if err != nil {
			return ok, err
		}
		r, err := b.Number()
		if err != nil {
			return ok, err
		}
		switch cmp {
		case GT:
			ok = l > r
		case GE:
			ok = l >= r
		case EQ:
			ok = l == r
		case NE:
			ok = l != r
		case LE:
			ok = l <= r
		case LT:
			ok = l < r
		}
		return ok, nil
	case KindList:
		switch cmp {
		case EQ:
			ok = a == b
		case NE:
			ok = a != b
		}
		return
	case KindObject:
		switch cmp {
		case EQ:
			ok = a == b
		case NE:
			ok = a != b
		}
		return
	case KindFunction:
		switch cmp {
		case EQ:
			ok = a == b
		case NE:
			ok = a != b
		}
		return
	default:
		return
	}
}

func binaryOPValues(a, b Value, op int) (Value, error) {
	if a.Kind() == KindList && op == ADD {
		lst, err := a.List()
		if err != nil {
			return nil, err
		}
		lst = append(lst, b)
		return ListValue(lst), nil
	}

	if a.Kind() != b.Kind() {
		goto returnError
	}

	switch a.Kind() {
	case KindString:
		goto handleString
	case KindBool:
		goto returnError
	case KindNumber:
		goto handleNumber
	case KindList:
		goto handleList
	case KindObject:
		goto handleObject
	}

returnError:
	return nil, &ErrType{opAdd, a.Kind().String(), conTO, b.Kind().String()}

handleString:
	{
		if op != ADD {
			goto returnError
		}

		l, err := a.String()
		if err != nil {
			return nil, err
		}
		r, err := b.String()
		if err != nil {
			return nil, err
		}
		return StringValue(l + r), nil
	}

handleNumber:
	{
		l, err := a.Number()
		if err != nil {
			return nil, err
		}
		r, err := b.Number()
		if err != nil {
			return nil, err
		}

		var v float64
		switch op {
		case ADD:
			v = l + r
		case SUB:
			v = l - r
		case MUL:
			v = l * r
		case DIV:
			v = l / r
		}
		return NumberValue(v), nil
	}
handleList:
	{
		lst, err := a.List()
		if err != nil {
			return nil, err
		}
		rlst, err := b.List()
		if err != nil {
			return nil, err
		}

		switch op {
		case ADD:
			lst = append(lst, rlst...)
		case SUB:
			items := map[Value]struct{}{}
			for _, v := range lst {
				items[v] = struct{}{}
			}
			for _, v := range rlst {
				delete(items, v)
			}
			lst = make([]Value, 0, len(items))
			for v := range items {
				lst = append(lst, v)
			}
		default:
			goto returnError
		}
		return ListValue(lst), err
	}

handleObject:
	{
		res := make(ObjectValue)

		// fast path if both values are ObjectValues
		if a, ok := a.(ObjectValue); ok {
			if b, ok := b.(ObjectValue); ok {
				for key, value := range a {
					res[key] = value
				}
				for key, value := range b {
					res[key] = value
				}
				return res, nil
			}
		}

		// fallback

		lObj, err := a.Object()
		if err != nil {
			return nil, err
		}
		rObj, err := b.Object()
		if err != nil {
			return nil, err
		}

		for _, key := range lObj.Keys() {
			res[key], _ = lObj.Key(key)
		}
		for _, key := range rObj.Keys() {
			res[key], _ = rObj.Key(key)
		}

		return res, nil
	}
}
