package html

import "testing"

func TestCompile(t *testing.T) {
	doc := `
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
	`

	CompileString(doc)
}
