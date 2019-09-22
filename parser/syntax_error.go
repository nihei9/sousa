package parser

import (
	"fmt"
	"strings"
)

// SyntaxError is the implementation of the error interface.
type SyntaxError struct {
	file     string
	position Position
	message  string
}

func (synErr *SyntaxError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "syntax error: %s\n", synErr.message)
	fmt.Fprintf(&b, "  %s:%v:%v\n", synErr.file, synErr.position.Line, synErr.position.Column)

	return b.String()
}
