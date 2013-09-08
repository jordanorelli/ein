package ein

import (
	"fmt"
	"io"
	"strings"
)

// the tokenReader interface provides a stream of tokens that can be "rewound"
// using the unread operation.
// TODO: make unread take an integer param instead of actual tokens; this
// current API is kinda dumb.
type tokenReader interface {
	// gets and consumes the next token
	next() token

	// gets and consumes the next n tokens
	nextn(int) []token

	// returns the value of the next token without consuming it
	peek() token

	// put some tokens back (rewind)
	unread(...token)

	// read all tokens until a specific token type comes up, returning all
	// intermediate tokens, up to but not including the terminal token.
	// Returns an error if the end of the token stream is reached without
	// finding a token of the provided type
	readUntil(tokenType) ([]token, error)
}

// type tokenChanReader wraps a channel of tokens, providing backup
// functionality so that tokens can be "unread".
type tokenChanReader struct {
	in      chan token
	history []token
}

func newTokenReader(c chan token) tokenReader {
	return &tokenChanReader{
		in:      c,
		history: make([]token, 0, 8),
	}
}

func (p *tokenChanReader) next() token {
	if len(p.history) > 0 {
		last := p.history[len(p.history)-1]
		p.history = p.history[0 : len(p.history)-1]
		return last
	}
	t, ok := <-p.in
	if !ok {
		panic("parsing a closed lex stream")
	}
	return t
}

func (p *tokenChanReader) nextn(n int) []token {
	tokens := make([]token, n)
	for i := 0; i < n; i++ {
		tokens[i] = p.next()
	}
	return tokens
}

func (p *tokenChanReader) unread(t ...token) {
	debugf("unread: %v\n", t)
	if len(t) == 0 {
		return
	}
	for i := len(t) - 1; i >= 0; i-- {
		p.history = append(p.history, t[i])
	}
}

func (p *tokenChanReader) peek() token {
	t := p.next()
	p.unread(t)
	return t
}

type parseError struct {
	message string
}

func (p parseError) Error() string {
	return p.message
}

// returns all tokens up to the next token type.  Errors if we read EOF before
// finding the correct token type.
func (p *tokenChanReader) readUntil(tt tokenType) ([]token, error) {
	out := make([]token, 0, 8)
	for {
		t := p.next()
		switch t.typ {
		case tokenEOF:
			if tt == tokenEOF {
				p.unread(t)
				return out, nil
			}
			msg := fmt.Sprintf("unexpected EOF while scanning for %v", tt)
			return nil, parseError{msg}
		case tt:
			p.unread(t)
			return out, nil
		default:
			out = append(out, t)
		}
	}
}

func matchInitial(tt tokenType) func([]token) bool {
	return func(tokens []token) bool {
		return len(tokens) > 0 && tokens[0].typ == tt
	}
}

func matchSequence(tt ...tokenType) func([]token) bool {
	return func(tokens []token) bool {
		if len(tt) != len(tokens) {
			return false
		}
		for i := 0; i < len(tt); i++ {
			if tokens[i].typ != tt[i] {
				return false
			}
		}
		return true
	}
}

func parseTag(tokens []token) (node, error) {
	debugf("parse tag: %v\n", tokens)
	switch {
	case matchSequence(tokenIdentifier)(tokens):
		return newIdentifierNode(tokens[0].val), nil
	case matchSequence(tokenEnd)(tokens):
		return newEndNode(), nil
	case matchSequence(tokenElse)(tokens):
		return newElseNode(), nil
	case matchInitial(tokenIf)(tokens):
		return newIfNode(tokens), nil
	default:
		return nil, fmt.Errorf("barf barf barf")
	}
}

func parse(source io.Reader) (node, error) {
	c := make(chan token)
	go lex(source, c)
	r, n := newTokenReader(c), newListNode(true)
	if err := n.parse(r); err != nil {
		return nil, err
	}
	return n, nil
}

func sparse(source string) (node, error) {
	return parse(strings.NewReader(source))
}

func isEnd(tokens []token) bool {
	return len(tokens) == 1 && tokens[0].typ == tokenEnd
}

type parsePredicate func(tokenReader) bool

func lookahead(name string, tokens ...token) parsePredicate {
	return func(r tokenReader) bool {
		debugf("lookahead %s sees: %v\n", name, tokens)
		upcoming := make([]token, 0, len(tokens))
		defer func() {
			r.unread(upcoming...)
		}()

		for i := 0; i < len(tokens); i++ {
			upcoming = append(upcoming, r.next())
			if upcoming[i] != tokens[i] {
				debugf("expected: %v got: %v at %d\n", tokens[i], upcoming[i], i)
				return false
			}
		}
		debugf("lookahead is returning true!\n")
		return true
	}
}

var (
	beforeEnd  = lookahead("beforeEnd", t_left, t_end, t_right)
	beforeElse = lookahead("beforeElse", t_left, t_else, t_right)
)
var parsePredicates = map[string]parsePredicate{
	"beforeEnd":  beforeEnd,
	"beforeElse": beforeElse,
}
