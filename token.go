package ein

import (
	"fmt"
)

type tokenType int

const (
	tokenInvalid    tokenType = iota // 0
	tokenError                       // 1
	tokenEOF                         // 2
	tokenPlaintext                   // 3
	tokenLeftMeta                    // 4
	tokenRightMeta                   // 5
	tokenIdentifier                  // 6
	tokenIf                          // 7
	tokenFor                         // 8
	tokenEnd                         // 9
	tokenElse                        // 10
)

var tokenNames = map[tokenType]string{
	tokenInvalid:    "invalid",
	tokenError:      "error",
	tokenEOF:        "EOF",
	tokenPlaintext:  "text",
	tokenLeftMeta:   "leftM",
	tokenRightMeta:  "rightM",
	tokenIdentifier: "ident",
	tokenIf:         "if",
	tokenFor:        "for",
	tokenEnd:        "end",
	tokenElse:       "else",
}

var (
	t_eof   = token{tokenEOF, "EOF"}
	t_left  = token{tokenLeftMeta, "{{"}
	t_right = token{tokenRightMeta, "}}"}
	t_end   = token{tokenEnd, "end"}
	t_else  = token{tokenElse, "else"}
)

func (t tokenType) String() string {
	return tokenNames[t]
}

type token struct {
	typ tokenType
	val string
}

func (t token) String() string {
    return fmt.Sprintf("{%s: %s}", t.typ, t.val)
}

// func (t token) String() string {
// 	switch t.typ {
// 	case tokenEOF:
// 		return "EOF"
// 	case tokenError:
// 		return t.val
// 	}
// 	// if len(t.val) > 10 {
// 	// 	return fmt.Sprintf("%v %.10q...", t.typ, t.val)
// 	// }
// 	return fmt.Sprintf("[%v %q]", t.typ, t.val)
// }
