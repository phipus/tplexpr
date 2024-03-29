// Code generated by "stringer -type TokenType -trimprefix Token"; DO NOT EDIT.

package tplexpr

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TokenValue-0]
	_ = x[TokenIdent-1]
	_ = x[TokenNumber-2]
	_ = x[TokenLeftParen-3]
	_ = x[TokenRightParen-4]
	_ = x[TokenDot-5]
	_ = x[TokenComma-6]
	_ = x[TokenEOF-7]
	_ = x[TokenString-8]
	_ = x[TokenArrow-9]
	_ = x[TokenDeclare-10]
	_ = x[TokenGT-11]
	_ = x[TokenGE-12]
	_ = x[TokenEQ-13]
	_ = x[TokenNE-14]
	_ = x[TokenLE-15]
	_ = x[TokenLT-16]
	_ = x[TokenAND-17]
	_ = x[TokenOR-18]
	_ = x[TokenADD-19]
	_ = x[TokenSUB-20]
	_ = x[TokenMUL-21]
	_ = x[TokenDIV-22]
	_ = x[TokenBlock-23]
	_ = x[TokenEndBlock-24]
	_ = x[TokenIf-25]
	_ = x[TokenThen-26]
	_ = x[TokenElse-27]
	_ = x[TokenElseIf-28]
	_ = x[TokenEndIf-29]
	_ = x[TokenFor-30]
	_ = x[TokenIn-31]
	_ = x[TokenDo-32]
	_ = x[TokenBreak-33]
	_ = x[TokenContinue-34]
	_ = x[TokenEndFor-35]
	_ = x[TokenInclude-36]
	_ = x[TokenDiscard-37]
	_ = x[TokenEndDiscard-38]
	_ = x[TokenObject-39]
	_ = x[TokenError-40]
}

const _TokenType_name = "ValueIdentNumberLeftParenRightParenDotCommaEOFStringArrowDeclareGTGEEQNELELTANDORADDSUBMULDIVBlockEndBlockIfThenElseElseIfEndIfForInDoBreakContinueEndForIncludeDiscardEndDiscardObjectError"

var _TokenType_index = [...]uint8{0, 5, 10, 16, 25, 35, 38, 43, 46, 52, 57, 64, 66, 68, 70, 72, 74, 76, 79, 81, 84, 87, 90, 93, 98, 106, 108, 112, 116, 122, 127, 130, 132, 134, 139, 147, 153, 160, 167, 177, 183, 188}

func (i TokenType) String() string {
	if i < 0 || i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
