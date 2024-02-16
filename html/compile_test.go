package html

import (
	"strings"
	"testing"

	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
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
		for name, value := range testCase.vars {
			c.Declare(name, value)
		}
		str, err := tplexpr.EvalString(&c, code)
		if err != nil {
			t.Error(err)
			continue
		}

		foundNode, err := html.Parse(strings.NewReader(str))
		if err != nil {
			t.Error(err)
			t.Log("Produced invalid html")
			continue
		}

		expectedNode, err := html.Parse(strings.NewReader(testCase.expected))
		if err != nil {
			t.Error(err)
			t.Log("Expected invalid html")
			continue
		}

		foundSB := strings.Builder{}
		err = html.Render(&foundSB, foundNode)
		if err != nil {
			t.Error(err)
			continue
		}

		expectedSB := strings.Builder{}
		err = html.Render(&expectedSB, expectedNode)
		if err != nil {
			t.Error(err)
			continue
		}

		found := foundSB.String()
		expected := expectedSB.String()

		if found != expected {
			t.Errorf("expected '%s', found '%s'", expected, found)
		}
	}
}
