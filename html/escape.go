package html

import "strings"

// Quote writes s to w surrounded by quotes. Normally it will use double
// quotes, but if s contains a double quote, it will use single quotes.
// It is used for writing the identifiers in a doctype declaration.
// In valid HTML, they can't contain both types of quotes.
func Quote(s string) string {
	q := byte('"')
	bytes := []byte{}

	if strings.Contains(s, `"`) {
		q = '\''
	}

	bytes = append(bytes, q)
	bytes = append(bytes, s...)
	bytes = append(bytes, q)

	return string(bytes)
}

// EscapeComment is like func escape but escapes its input bytes less often.
// Per https://github.com/golang/go/issues/58246 some HTML comments are (1)
// meaningful and (2) contain angle brackets that we'd like to avoid escaping
// unless we have to.
//
// "We have to" includes the '&' byte, since that introduces other escapes.
//
// It also includes those bytes (not including EOF) that would otherwise end
// the comment. Per the summary table at the bottom of comment_test.go, this is
// the '>' byte that, per above, we'd like to avoid escaping unless we have to.
//
// Studying the summary table (and T actions in its '>' column) closely, we
// only need to escape in states 43, 44, 49, 51 and 52. State 43 is at the
// start of the comment data. State 52 is after a '!'. The other three states
// are after a '-'.
//
// Our algorithm is thus to escape every '&' and to escape '>' if and only if:
//   - The '>' is after a '!' or '-' (in the unescaped data) or
//   - The '>' is at the start of the comment data (after the opening "<!--").
func EscapeComment(s string) string {
	// When modifying this function, consider manually increasing the
	// maxSuffixLen constant in func TestComments, from 6 to e.g. 9 or more.
	// That increase should only be temporary, not committed, as it
	// exponentially affects the test running time.

	if len(s) == 0 || !strings.ContainsAny(s, "&>") {
		return s
	}

	bytes := []byte{}

	// Loop:
	//   - Grow j such that s[i:j] does not need escaping.
	//   - If s[j] does need escaping, output s[i:j] and an escaped s[j],
	//     resetting i and j to point past that s[j] byte.
	i := 0
	for j := 0; j < len(s); j++ {
		escaped := ""
		switch s[j] {
		case '&':
			escaped = "&amp;"

		case '>':
			if j > 0 {
				if prev := s[j-1]; (prev != '!') && (prev != '-') {
					continue
				}
			}
			escaped = "&gt;"

		default:
			continue
		}

		if i < j {
			bytes = append(bytes, s[i:j]...)
		}
		bytes = append(bytes, escaped...)
		i = j + 1
	}

	if i < len(s) {
		bytes = append(bytes, s[i:]...)
	}
	return string(bytes)
}
