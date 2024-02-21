package html

import (
	"io/fs"
	"os"
	"path"
	"strings"
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
	fsys := os.DirFS("testdata")
	store, err := tplexpr.BuildStore().
		AddPlugin(&Plugin{}).
		AddFS(fsys, "*.test.html", "*.template.html").
		Build()
	if err != nil {
		t.Error(err)
		return
	}

	matches, err := fs.Glob(fsys, "*.test.html")
	if err != nil {
		t.Error(err)
		return
	}

	for _, fileName := range matches {
		t.Logf("eval template %s", fileName)

		sb := strings.Builder{}
		err = store.Render(&sb, fileName, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		str := sb.String()

		os.WriteFile(path.Join("testdata/tmp", fileName), []byte(str), 0644)

		resultFileName := strings.Replace(fileName, ".test.html", ".result.html", 1)

		expectedBytes, err := os.ReadFile(path.Join("testdata", resultFileName))
		if err != nil {
			t.Error(err)
			continue
		}
		// convert to linux line endings
		expected := strings.ReplaceAll(string(expectedBytes), "\r\n", "\n")

		if str != string(expected) {
			t.Error("expected:")
			t.Error(string(expected))
			t.Error("found:")
			t.Error(str)
		}
	}
}
