package ein

import (
	"testing"
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

func (test lexTest) run(t *testing.T) {
	tokens := slexAll(test.input)
	if !equal(tokens, test.tokens) {
		t.Errorf("LEX %s:\n\tgot:\n\t\t%v\n\texpected:\n\t\t%v\n", test.name, tokens, test.tokens)
	}
}

var lexTests = []lexTest{
	{"empty", "", []token{t_eof}},
	{"spaces", " \t\n", []token{
		{tokenPlaintext, " \t\n"},
		t_eof,
	}},
	{"plain text", "this is some plain text", []token{
		{tokenPlaintext, "this is some plain text"},
		t_eof,
	}},
	{"left meta", "{{", []token{
		t_left,
		t_eof,
	}},
	{"right meta", "}}", []token{
		{tokenError, "unexpected right meta in lexPlaintext"},
		t_eof,
	}},
	{"simple tag", "opening text {{tag_identifier}} closing text", []token{
		{tokenPlaintext, "opening text "},
		t_left,
		{tokenIdentifier, "tag_identifier"},
		t_right,
		{tokenPlaintext, " closing text"},
		t_eof,
	}},
	{"end tag", "{{end}}", []token{
		t_left,
		t_end,
		t_right,
		t_eof,
	}},
	// should an empty tag be an error here, or should that error in parse?
	{"empty tag", "{{}}", []token{
		t_left,
		t_right,
		t_eof,
	}},
}

func equal(t1, t2 []token) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i := 0; i < len(t1); i++ {
		if t1[i] != t2[i] {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		test.run(t)
	}
}
