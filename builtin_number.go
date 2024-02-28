package tplexpr

import (
	"math"
)

func mapNumber1(f func(float64) float64) FuncValue {
	return func(args Args) (Value, error) {
		nr, err := args.GetDefault(0, Zero).Number()
		return NumberValue(f(nr)), err
	}
}

func mapNumber2(f func(a, b float64) float64) FuncValue {
	return func(args Args) (Value, error) {
		a, err := args.GetDefault(0, Zero).Number()
		if err != nil {
			return nil, err
		}
		b, err := args.GetDefault(1, Zero).Number()
		return NumberValue(f(a, b)), err
	}
}

func mapNumberBool1(f func(float64) bool) FuncValue {
	return func(args Args) (Value, error) {
		nr, err := args.GetDefault(0, Zero).Number()
		return BoolValue(f(nr)), err
	}
}

var numberBuiltins = map[string]Value{
	"abs":         mapNumber1(math.Abs),
	"acos":        mapNumber1(math.Acos),
	"acosh":       mapNumber1(math.Acosh),
	"asin":        mapNumber1(math.Asin),
	"asinh":       mapNumber1(math.Asinh),
	"atan":        mapNumber1(math.Atan),
	"atan2":       mapNumber2(math.Atan2),
	"atanh":       mapNumber1(math.Atanh),
	"cbrt":        mapNumber1(math.Cbrt),
	"ceil":        mapNumber1(math.Ceil),
	"cos":         mapNumber1(math.Cos),
	"cosh":        mapNumber1(math.Cosh),
	"exp":         mapNumber1(math.Exp),
	"exp2":        mapNumber1(math.Exp2),
	"floor":       mapNumber1(math.Floor),
	"gamma":       mapNumber1(math.Gamma),
	"hypot":       mapNumber2(math.Hypot),
	"inf":         FuncValue(BuiltinIsInf),
	"isInf":       FuncValue(BuiltinIsInf),
	"isNaN":       mapNumberBool1(math.IsNaN),
	"log":         mapNumber1(math.Log),
	"log10":       mapNumber1(math.Log10),
	"log2":        mapNumber1(math.Log2),
	"max":         FuncValue(BuiltinMax),
	"min":         FuncValue(BuiltinMin),
	"mod":         mapNumber2(math.Mod),
	"NaN":         NumberValue(math.NaN()),
	"pow":         mapNumber2(math.Pow),
	"round":       mapNumber1(math.Round),
	"roundToEven": mapNumber1(math.RoundToEven),
	"sin":         mapNumber1(math.Sin),
	"sinh":        mapNumber1(math.Sinh),
	"sqrt":        mapNumber1(math.Sqrt),
	"tan":         mapNumber1(math.Tan),
	"tanh":        mapNumber1(math.Tanh),
	"trunc":       mapNumber1(math.Trunc),
}

func BuiltinInf(args Args) (Value, error) {
	sign, err := args.GetDefault(0, Zero).Number()
	return NumberValue(math.Inf(int(sign))), err
}

func BuiltinIsInf(args Args) (Value, error) {
	nr, err := args.GetDefault(0, Zero).Number()
	if err != nil {
		return nil, err
	}
	sign, err := args.GetDefault(1, Zero).Number()
	return BoolValue(math.IsInf(nr, int(sign))), err
}

func BuiltinMax(args Args) (v Value, err error) {
	if args.Len() <= 0 {
		return Nil, nil
	}

	nrs := make([]float64, args.Len())

	for i := range nrs {
		nrs[i], err = args.Get(i).Number()
		if err != nil {
			return
		}
	}

	max := nrs[0]
	for i := 1; i < len(nrs); i++ {
		if nrs[i] > max {
			max = nrs[i]
		}
	}
	return NumberValue(max), nil
}

func BuiltinMin(args Args) (v Value, err error) {
	if args.Len() <= 0 {
		return Nil, nil
	}

	nrs := make([]float64, args.Len())

	for i := range nrs {
		nrs[i], err = args.Get(i).Number()
		if err != nil {
			return
		}
	}

	min := nrs[0]
	for i := 1; i < len(nrs); i++ {
		if nrs[i] < min {
			min = nrs[i]
		}
	}
	return NumberValue(min), nil
}

func AddNumberBuiltins(c *Context) {
	for name, value := range numberBuiltins {
		c.Declare(name, value)
	}
}
