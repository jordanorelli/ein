package ein

import (
	"reflect"
	"testing"
)

type parseTest struct {
	name  string
	input string
	// if failurey, should produce an error with this message
	err string
	// if successful, should produce this node
	output node
}

func (test parseTest) run(t *testing.T) {
	n, err := sparse(test.input)
	if err != nil {
		if err.Error() != test.err {
			t.Errorf("PARSE %s FAIL ERROR:\n\tgot:\n\t\t%v\n\texpected:\n\t\t%v\n",
				test.name, err.Error(), test.err)
		}
	} else {
		if !reflect.DeepEqual(n, test.output) {
			t.Errorf("PARSE %s:\n\tgot:\n\t\t%v\n\texpected:\n\t\t%v\n",
				test.name, n, test.output)
		}
	}
}

var parseTests = []parseTest{
	{"empty", "", "", newListNode(true)},
	{"plain text", "this is some plain text", "", &listNode{
		nodeType: nodeList,
		root:     true,
		children: children{
			plaintextNode{
				nodeType: nodePlaintext,
				text:     "this is some plain text",
			},
		},
	}},
	{"variable", "{{x}}", "", &listNode{
		nodeType: nodeList,
		root:     true,
		children: children{
			identifierNode{
				nodeType: nodeIdentifier,
				name:     "x",
			},
		},
	}},
	{"plain text + variable", "opening {{x}} closing", "", &listNode{
		nodeType: nodeList,
		root:     true,
		children: children{
			plaintextNode{
				nodeType: nodePlaintext,
				text:     "opening ",
			},
			identifierNode{
				nodeType: nodeIdentifier,
				name:     "x",
			},
			plaintextNode{
				nodeType: nodePlaintext,
				text:     " closing",
			},
		},
	}},
	{"if statement", "{{if x}}tales of us{{end}}", "", &listNode{
		nodeType: nodeList,
		root:     true,
		children: children{
			&ifNode{
				nodeType: nodeIf,
				cond: []token{
					{tokenIdentifier, "x"},
				},
				trueBranch: &listNode{
					nodeType: nodeList,
					children: children{
						plaintextNode{
							nodeType: nodePlaintext,
							text:     "tales of us",
						},
					},
					terminators: []string{"beforeEnd", "beforeElse"},
				},
			},
		},
	}},
	{"if statement with else condition", "{{if x}}tales of us{{else}}voicething{{end}}", "", &listNode{
		nodeType: nodeList,
		root:     true,
		children: children{
			&ifNode{
				nodeType: nodeIf,
				cond: []token{
					{tokenIdentifier, "x"},
				},
				trueBranch: &listNode{
					nodeType: nodeList,
					children: children{
						plaintextNode{
							nodeType: nodePlaintext,
							text:     "tales of us",
						},
					},
					terminators: []string{"beforeEnd", "beforeElse"},
				},
				falseBranch: &listNode{
					nodeType: nodeList,
					children: children{
						plaintextNode{
							nodeType: nodePlaintext,
							text:     "voicething",
						},
					},
					terminators: []string{"beforeEnd"},
				},
			},
		},
	}},
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		test.run(t)
	}
}
