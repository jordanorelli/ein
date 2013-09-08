package ein

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	in  *bufio.Reader
	out chan token // channel of scanned tokens
	buf []rune
}

func newLexer(in io.Reader, out chan token) *lexer {
	l := &lexer{
		in:  bufio.NewReader(in),
		out: out,
		buf: make([]rune, 0, 32),
	}
	return l
}

func (l *lexer) next() rune {
	r, _, err := l.in.ReadRune()
	switch err {
	case nil:
	case io.EOF:
		return eof
	default:
		l.errorf("lex error in next: %v", err)
		return eof
	}
	l.buf = append(l.buf, r)
	return r
}

func (l *lexer) discard() {
	if len(l.buf) >= 1 {
		l.buf = l.buf[0 : len(l.buf)-1]
	}
}

func (l *lexer) backup() {
	l.discard()
	err := l.in.UnreadRune()
	if err != nil {
		l.errorf("lex error in backup: %v", err)
	}
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) clearBuf() {
	l.buf = l.buf[0:0]
}

func (l *lexer) emit(t tokenType) {
	defer l.clearBuf()
	tok := token{t, string(l.buf)}
	debugf("lex out: %v\n", tok)
	l.out <- tok
}

func (l *lexer) emitError(format string, args ...interface{}) {
	debugf("lex error: "+format, args...)
	l.out <- token{tokenError, fmt.Sprintf(format, args...)}
}

func (l *lexer) fatalf(format string, args ...interface{}) stateFn {
	l.emitError(format, args...)
	return nil
}

func (l *lexer) skipUntil(good string) {
	for {
		r := l.next()
		if r == eof || strings.ContainsRune(good, r) {
			l.backup()
			return
		}
	}
}

var keywords = map[string]tokenType{
	"if":   tokenIf,
	"else": tokenElse,
	"end":  tokenEnd,
}

func (l *lexer) identifier() {
	if t, ok := keywords[string(l.buf)]; ok {
		l.emit(t)
	} else {
		l.emit(tokenIdentifier)
	}
}

func lexIdentifier(l *lexer) stateFn {
	switch r := l.next(); {
	case unicode.IsLetter(r), unicode.IsDigit(r), r == '_':
		return lexIdentifier
	default:
		l.backup()
		l.identifier()
		return lexLeftMeta
	}
}

func lexLeftMeta(l *lexer) stateFn {
	switch r := l.next(); {
	case r == '}':
		if l.peek() == '}' {
			l.next()
			l.emit(tokenRightMeta)
		}
		return lexPlaintext
	case unicode.IsLetter(r):
		return lexIdentifier
	case unicode.IsSpace(r):
		l.discard()
		return lexLeftMeta
	case r == eof:
		return nil
	default:
		return lexLeftMeta
	}
}

func lexPlaintext(l *lexer) stateFn {
	switch l.next() {
	case '{':
		if l.peek() == '{' {
			l.next()
			l.discard()
			l.discard()
            if len(l.buf) > 0 {
                l.emit(tokenPlaintext)
            }
			// this is pretty janky, but it's either this or create my own
			// infinitely unreadable io.Reader implementation, which is kinda a
			// waste of time for this one very constrained use case.
			l.buf = append(l.buf, '{', '{')
			l.emit(tokenLeftMeta)
			return lexLeftMeta
		}
		return lexPlaintext
    case '}':
        if l.peek() == '}' {
            l.emitError("unexpected right meta in lexPlaintext")
            return nil
        }
        return lexPlaintext
	case eof:
        if len(l.buf) > 0 {
            l.emit(tokenPlaintext)
        }
		return nil
	default:
		return lexPlaintext
	}
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.emitError(format, args...)
	l.skipUntil("\n\r")
	l.next()
	l.clearBuf()
	return lexPlaintext
}

func (l *lexer) done() {
	l.out <- token{tokenEOF, "EOF"}
	close(l.out)
}

func lex(in io.Reader, c chan token) {
	lexer := newLexer(in, c)
	defer lexer.done()

	for fn := lexPlaintext; fn != nil; {
		fn = fn(lexer)
	}
}

func slex(in string, c chan token) {
    lex(strings.NewReader(in), c)
}

func lexAll(in io.Reader) []token {
    c := make(chan token)
    go lex(in, c)
    tokens := make([]token, 0, 32)
    for token := range c {
        tokens = append(tokens, token)
    }
    return tokens
}

func slexAll(in string) []token {
    return lexAll(strings.NewReader(in))
}
