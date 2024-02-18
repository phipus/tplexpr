package html

import (
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/phipus/tplexpr"
)

func TestCompile(t *testing.T) {
	type testCase struct {
		name     string
		doc      string
		expected string
		vars     tplexpr.Vars
	}

	testCases := []testCase{
		{
			name: "Test Plain",
			doc: `
				<!DOCTYPE html>
				<html>
					<head>
						<title>Hello</title>
					</head>
					<body>
						<!-- Comment -->
						Hello <b>World</b>
						<h3>Heading</h3>
						<p>paragraph text</p>
					</body>
				</html>
			`,
			expected: `
				<!DOCTYPE html>
				<html>
					<head>
						<title>Hello</title>
					</head>
					<body>
						<!-- Comment -->
						Hello <b>World</b>
						<h3>Heading</h3>
						<p>paragraph text</p>
					</body>
				</html>
			`,
		},
		{
			name: "Substitute 1",
			doc: `
				<!DOCTYPE html>
				<html>
					<head>
						<title>Hello $name</title>
					</head>
					<body>
						<h1>Hello $name</h1>
					</body>
				</html>
			`,
			expected: `
				<!DOCTYPE html>
				<html>
					<head>
						<title>Hello World</title>
					</head>
					<body>
						<h1>Hello World</h1>
					</body>
				</html>
			`,
			vars: tplexpr.Vars{"name": tplexpr.S("World")},
		},
		{
			name: "Iterate (tx-for)",
			doc: `
				<ul>
					<tx-for var="i" expr="${range(1, 6)}">
						<li>$i</li>
					</tx-for>
				</ul>
			`,
			expected: `
				<ul>
					
						<li>1</li>
					
						<li>2</li>
					
						<li>3</li>
					
						<li>4</li>
					
						<li>5</li>
					
				</ul>
			`,
		},
	}

	for _, testCase := range testCases {
		t.Logf("run testcase %s", testCase.name)

		ctx := tplexpr.NewCompileContext()
		err := CompileString(testCase.doc, &ctx, tplexpr.CompileEmit)
		if err != nil {
			t.Error(err)
			continue
		}

		code, c := ctx.Compile()
		tplexpr.AddBuiltins(&c)
		for name, value := range testCase.vars {
			c.Declare(name, value)
		}
		str, err := tplexpr.EvalString(&c, code)
		if err != nil {
			t.Error(err)
			continue
		}

		if str != testCase.expected {
			t.Errorf("expected '%s', found '%s'", testCase.expected, str)
		}
	}
}

func TestCompileFiles(t *testing.T) {
	fsys, err := fs.Sub(os.DirFS("testdata"), "templates")
	if err != nil {
		t.Error(err)
		return
	}

	ctx := tplexpr.NewCompileContext()
	err = CompileGlobFS(fsys, "*.html", &ctx)
	if err != nil {
		t.Error(err)
		return
	}
	_, c := ctx.Compile()
	tplexpr.AddBuiltins(&c)

	templates := []string{
		"test1.html",
	}

	for _, tpl := range templates {
		t.Logf("eval template %s", tpl)
		str, err := c.EvalTemplateString(tpl, nil)
		if err != nil {
			t.Error(err)
			continue
		}

		os.WriteFile(path.Join("testdata/tmp", tpl), []byte(str), 0644)

		expected, err := os.ReadFile(path.Join("testdata/results", tpl))
		if err != nil {
			t.Error(err)
			continue
		}
		if str != string(expected) {
			t.Error("expected:")
			t.Error(string(expected))
			t.Error("found:")
			t.Error(str)
		}
	}
}
