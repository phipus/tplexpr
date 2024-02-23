package tplexpr

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

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
			vars: Vars{
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
		{`${block(tpl, name)}Hello $name${endblock}`, "", nil},
		{`${block(tpl, name)}Hello $name${endblock}${tpl("World")}`, "Hello World", nil},
		{`${if false then "A" elseif true then "B" else "C" endif}`, "B", nil},
		{`${if true then "A" elseif true then "B" else "C" endif}`, "A", nil},
		{`${if true then "A" elseif false then "B" else "C" endif}`, "A", nil},
		{`${if false then "A" elseif false then "B" else "C" endif}`, "C", nil},
		{`${range(1, 6)}`, `1 2 3 4 5`, nil},
		{`${range(6)}`, `0 1 2 3 4 5`, nil},
		{`${range(0, -1, -1)}`, `0`, nil},
		{`${range(0, -1)}`, ``, nil},
		{`${range(2, 0, -1)}`, `2 1`, nil},
		{`${declare(x, object(a => 1, b => 2))}${x.b}${x.a}`, "21", nil},
		{`${declare(x, object(a => 1)) declare(y, object(x, b => 2))}${y.a y.b}`, "12", nil},
		{`${declare(x, list(1)) declare(y, x.append(2, 3)) y}`, "1 2 3", nil},
		{`${declare(x, list(1)) declare(y, list(2, 3)) x.extend(y)}`, "1 2 3", nil},
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

func TestEvalTemplate(t *testing.T) {
	evalTest("include 1").
		Template("baseFuncs",
			`${discard}
				${block(hello, name)}Hello $name${endblock}
				${declare(greet, (name) => "Hello $name")}
			${enddiscard}`,
		).
		Template("main",
			`${include("baseFuncs")%} ${hello("World")} ${greet("Sina")}`,
		).
		Eval(t, "main",
			`Hello World Hello Sina`,
		)
}

type evalTestImpl struct {
	name      string
	templates map[string]string
	vars      Vars
}

func evalTest(name string) *evalTestImpl {
	return &evalTestImpl{name: name, templates: map[string]string{}, vars: map[string]Value{}}
}

func (t *evalTestImpl) Template(name string, value string) *evalTestImpl {
	t.templates[name] = value
	return t
}

func (t *evalTestImpl) Var(name string, value Value) *evalTestImpl {
	t.vars[name] = value
	return t
}

func (t *evalTestImpl) VarString(name string, value string) *evalTestImpl {
	return t.Var(name, StringValue(value))
}

func (t *evalTestImpl) Eval(tt *testing.T, templateName string, expectedResult string) {
	tt.Logf("run eval test %s", t.name)
	cc := NewCompileContext()
	for name, value := range t.templates {
		err := cc.ParseTemplate(name, []byte(value))
		if err != nil {
			tt.Error(err)
			return
		}
	}

	_, c := cc.Compile()
	result, err := c.EvalTemplateString(templateName, t.vars)
	if err != nil {
		tt.Error(err)
		return
	}

	if result != expectedResult {
		tt.Errorf("expected '%s', got '%s'", expectedResult, result)
	}
}

func TestEvalFile(t *testing.T) {
	fsys := os.DirFS("testdata")
	glob := "*.test.txt"
	matches, err := fs.Glob(fsys, glob)
	if err != nil {
		t.Error(err)
		return
	}
	store, err := BuildStore().
		AddFS(fsys, glob).
		Build()
	if err != nil {
		t.Error(err)
		return
	}

	type subReflect struct {
		S string
		B bool
		I int
	}

	type reflect struct {
		Numbers []int
		Floats  []float64
		S       string
		X       interface{}
		Opt     *reflect
		Sub     subReflect
	}

	r := Reflect(reflect{
		Numbers: []int{1, 2, 3, 4},
		Floats:  []float64{0.25, 0.5, 0.75, 1},
		S:       "Hello",
		X:       &subReflect{S: "World", B: false, I: 42},
		Opt:     nil,
		Sub: subReflect{
			S: "Sub reflect Value",
			B: true,
			I: 48,
		},
	})

	vars := BuildVars().
		Set("lst", L{N(1), N(2), S("one"), S("two")}).
		SetString("s", "Hello World").
		SetString("q", "Q").
		Set("r", r)

	for _, fileName := range matches {
		t.Logf("Test with file %s", fileName)
		result, err := fs.ReadFile(fsys, strings.TrimSuffix(fileName, ".test.txt")+".result.txt")
		if err != nil {
			t.Error(err)
			return
		}
		sb := strings.Builder{}
		err = store.Render(&sb, fileName, vars.Build())
		if err != nil {
			t.Error(err)
			continue
		}

		str := sb.String()
		if str != string(result) {
			t.Error("expected:")
			t.Error(string(result))
			t.Error("found:")
			t.Error(str)
		}
	}
}
