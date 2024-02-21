package html

import (
	"net/url"
	"strings"

	"github.com/phipus/tplexpr"
)

func BuiltinBuildQueryParams(args tplexpr.Args) (tplexpr.Value, error) {
	obj, err := args.Arg(0).Object()
	if err != nil {
		return nil, err
	}

	sb := strings.Builder{}
	for i, key := range obj.Keys() {
		if i != 0 {
			sb.WriteByte('&')
		}
		value, _ := obj.Key(key)
		strValue := ""
		if value != nil {
			strValue, err = value.String()
			if err != nil {
				return nil, err
			}
		}

		sb.WriteString(key)
		sb.WriteByte('=')
		sb.WriteString(url.QueryEscape(strValue))
	}

	return tplexpr.StringValue(sb.String()), nil
}

func BuiltinQueryEscape(args tplexpr.Args) (tplexpr.Value, error) {
	str, err := args.Arg(0).String()
	if err != nil {
		return nil, err
	}
	return tplexpr.StringValue(url.QueryEscape(str)), nil
}

func BuiltinPathEscape(args tplexpr.Args) (tplexpr.Value, error) {
	str, err := args.Arg(0).String()
	if err != nil {
		return nil, err
	}
	return tplexpr.StringValue(url.PathEscape(str)), nil
}
