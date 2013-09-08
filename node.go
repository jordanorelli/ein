package ein

import (
	"fmt"
    "bytes"
)

type nodeType int

func (n nodeType) Type() nodeType {
	return n
}

const (
	nodeInvalid nodeType = iota
	nodeIdentifier
	nodePlaintext
	nodeList
	nodeIf
	nodeBlock
	nodeEnd
	nodeElse
)

var nodeNames = map[nodeType]string{
	nodeInvalid:    "invalid",
	nodeIdentifier: "identifier",
	nodePlaintext:  "plaintext",
	nodeList:       "list",
	nodeIf:         "if",
	nodeBlock:      "block",
	nodeEnd:        "end",
	nodeElse:       "else",
}

type node interface {
	Type() nodeType
	Children() []node
    String() string
	parse(tokenReader) error
}

type nilParser int

func (p nilParser) parse(r tokenReader) error {
	return nil
}

type childless int

func (c childless) Children() []node { return nil }

// plaintext node holds plain text.  Nothing special here; it doesn't actually
// do anything.
type plaintextNode struct {
	nodeType
	childless
	nilParser
	text string
}

func newPlaintextNode(text string) node {
	return plaintextNode{nodeType: nodePlaintext, text: text}
}

func (n plaintextNode) String() string {
    return fmt.Sprintf("[text: %q]", n.text)
}

// identifier node holds an identifier name, to be retrieved from an
// environment at execution time.
type identifierNode struct {
	nodeType
	childless
	nilParser
	name string
}

func newIdentifierNode(name string) node {
	return identifierNode{nodeType: nodeIdentifier, name: name}
}

func (n identifierNode) String() string {
	return fmt.Sprintf("[ident: %q]", n.name)
}

type children []node

func (c *children) push(child node) {
	*c = append(*c, child)
}

func (c children) Children() []node {
	if c == nil {
		return nil
	}
	return []node(c)
}

type listNode struct {
	nodeType
	children
	root        bool
	terminators []string
}

func (l listNode) String() string {
    var buf bytes.Buffer
    for _, n := range l.children {
        buf.WriteString(n.String())
    }
    return fmt.Sprintf("[list (%t, %d, %q): [%s] %q]", l.root, len(l.children), l.terminators, buf.String())
}

func newListNode(root bool, terminators ...string) *listNode {
	return &listNode{
		nodeType:    nodeList,
		children:    make(children, 0, 4),
		root:        root,
		terminators: terminators,
	}
}

func (n *listNode) parse(r tokenReader) error {
	for !n.atEnd(r) {
		switch t := r.next(); t.typ {
		case tokenLeftMeta:
			tokens, err := r.readUntil(tokenRightMeta)
			if err != nil {
				return err
			}
			child, err := parseTag(tokens)
			if err != nil {
				return err
			}
			tokens = append(tokens, r.next())
			if err := child.parse(r); err != nil {
				return err
			}
			n.push(child)
		case tokenPlaintext:
			n.push(newPlaintextNode(t.val))
		case tokenError:
			return fmt.Errorf("unexpected error token: %v", t)
		case tokenEOF:
			if n.root {
				return nil
			}
		default:
			return fmt.Errorf("unexpected %v token in parse(): %v", t.typ, t)
		}
	}
	return nil
}

func (n *listNode) atEnd(r tokenReader) bool {
	for _, name := range n.terminators {
        fn, ok := parsePredicates[name]
        if !ok {
            panic("invalid parsePredicate name: " + name)
        }
		if fn(r) {
            debugf("hit listNode end!\n")
			return true
		}
        debugf("listNode end terminator %v didn't pass...\n", name)
	}
	return false
}

type ifNode struct {
	nodeType
	cond        []token
	trueBranch  *listNode
	falseBranch *listNode
}

func (n ifNode) Children() []node {
	return nil
}

func (n ifNode) String() string {
    return fmt.Sprintf("[if cond:%v true:%s false:%s]", n.cond, n.trueBranch, n.falseBranch)
}

func newIfNode(tokens []token) *ifNode {
	return &ifNode{nodeType: nodeIf, cond: tokens[1:]}
}

func (n *ifNode) parse(r tokenReader) error {
	n.trueBranch = newListNode(false, "beforeEnd", "beforeElse")
	if err := n.trueBranch.parse(r); err != nil {
		return err
	}
	switch {
	case beforeEnd(r):
        r.nextn(3)
		return nil
	case beforeElse(r):
        r.nextn(3)
		n.falseBranch = newListNode(false, "beforeEnd")
        if err := n.falseBranch.parse(r); err != nil {
            return err
        }
        r.nextn(3)
        return nil
	default:
		panic("PUKE PUKE PUKE")
	}
}

type endNode struct {
	nodeType
	childless
	nilParser
}

func newEndNode() *endNode {
	return &endNode{nodeType: nodeEnd}
}

func (n endNode) String() string {
	return "[end]"
}

type elseNode struct {
	nodeType
	childless
	nilParser
}

func (n elseNode) String() string {
	return "[else]"
}

func newElseNode() *elseNode {
	return &elseNode{nodeType: nodeElse}
}

func matchNodes(left, right node) bool {
	if left.Type() != right.Type() {
		return false
	}
	c_left, c_right := left.Children(), right.Children()
	if len(c_left) != len(c_right) {
		return false
	}
	for i := range c_left {
		if !matchNodes(c_left[i], c_right[i]) {
			return false
		}
	}
	fmt.Println("match: ", left, right)
	return true
}
