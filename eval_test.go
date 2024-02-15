package tplexpr

import "testing"

func TestEval(t *testing.T) {
	type testCase struct {
		input  string
		result string
		vars   map[string]Value
	}

	testCases := []testCase{
		{"Hello World", "Hello World", nil},
		{"Hello $name", "Hello Peter", map[string]Value{"name": StringValue("Peter")}},
		{"Hello ${name}", "Hello Peter", map[string]Value{"name": StringValue("Peter")}},
		{
			input:  `${v.map((name) => "Hello $name")}`,
			result: "Hello Peter Hello Sina",
			vars:   map[string]Value{"v": ListValue{StringValue("Peter"), StringValue("Sina")}},
		},
		{
			input:  `<ul>${items.map((x) => "<li>${x.value}</li>").join()}</ul>`,
			result: "<ul><li>Hello</li><li>World</li></ul>",
			vars: map[string]Value{
				"items": L{
					O{"value": S("Hello")},
					O{"value": S("World")},
				},
			},
		},
		{}, // empty value
		{`${"" || ""}`, "", nil},
		{`${"" || "Hello"}`, "Hello", nil},
		{`${"Hello" || "World"}`, "Hello", nil},
		{`${"Hello" || ""}`, "Hello", nil},
		{`${"" && ""}`, "", nil},
		{`${"" && "Hello"}`, "", nil},
		{`${"Hello" && "World"}`, "World", nil},
		{`${"Hello" && ""}`, "", nil},
		{`${"1".number() + "2".number()}`, "3", nil},
		{`${"1" + "2"}`, "12", nil},
		{`${2 + 2 / 4}`, "2.5", nil},
		{`${declare(name, "Sina")}Hello $name`, "Hello Sina", nil},
		{`${declare(name, "Sina") "Hello $name"}`, "Hello Sina", nil},
		{`${declare(name, "Sina") %}  Hello $name`, "Hello Sina", nil},
		{`${list("a", "b", "c", "1")}`, "a b c 1", nil},
		{`${for x in list("1", "2", "3") do "$x" endfor}`, "123", nil},
		{`${template(tpl, name)}Hello $name${endtemplate}`, "", nil},
		{`${template(tpl, name)}Hello $name${endtemplate}${tpl("World")}`, "Hello World", nil},
		{`${if false then "A" elseif true then "B" else "C" endif}`, "B", nil},
		{`${if true then "A" elseif true then "B" else "C" endif}`, "A", nil},
		{`${if true then "A" elseif false then "B" else "C" endif}`, "A", nil},
		{`${if false then "A" elseif false then "B" else "C" endif}`, "C", nil},
	}

	for i := range testCases {
		testCase := &testCases[i]

		t.Logf("Eval %s", testCase.input)

		p := NewParser([]byte(testCase.input))
		n, err := p.Parse()
		if err != nil {
			t.Error(err)
			continue
		}

		cc := NewCompileContext()
		err = n.Compile(&cc, CompileEmit)
		if err != nil {
			t.Error(err)
			continue
		}

		code, c := cc.Compile()
		AddBuiltins(&c)
		for name, value := range testCase.vars {
			c.Assign(name, value)
		}

		result, err := EvalString(&c, code)
		if err != nil {
			t.Error(err)
			continue
		}

		if result != testCase.result {
			t.Errorf("expected '%s', got '%s'", testCase.result, result)
		}
	}
}
