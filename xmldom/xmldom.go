package xmldom

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type NodeKind int

const (
	ElementNode NodeKind = iota
	TextNode
	CommentNode
	DirectiveNode
	ProcInstNode
)

type Node struct {
	Kind     NodeKind
	Start    xml.StartElement
	End      xml.EndElement
	Children []*Node
	Data     []byte
	PI       *xml.ProcInst
}

func NewElement(local string) *Node {
	name := xml.Name{Local: local}
	return &Node{
		Kind:  ElementNode,
		Start: xml.StartElement{Name: name},
		End:   xml.EndElement{Name: name},
	}
}

func NewText(data string) *Node {
	return &Node{Kind: TextNode, Data: []byte(data)}
}

func ParseXML(r io.Reader) ([]*Node, error) {
	dec := xml.NewDecoder(r)
	var root = &Node{Kind: ElementNode, Start: xml.StartElement{Name: xml.Name{Local: "TMP"}}}
	stack := []*Node{root}

	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &Node{Kind: ElementNode, Start: t}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) == 1 {
				return nil, fmt.Errorf("xmldom.ParseXML: unexpected end element: %v", t.Name)
			}
			cur := stack[len(stack)-1]
			cur.End = t
			stack = stack[:len(stack)-1]
		case xml.CharData:
			b := append([]byte(nil), t...)
			n := &Node{Kind: TextNode, Data: b}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
		case xml.Comment:
			b := append([]byte(nil), t...)
			n := &Node{Kind: CommentNode, Data: b}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
		case xml.Directive:
			b := append([]byte(nil), t...)
			n := &Node{Kind: DirectiveNode, Data: b}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
		case xml.ProcInst:
			pi := t
			n := &Node{Kind: ProcInstNode, PI: &pi}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, n)
		}
	}
	if len(stack) != 1 {
		return nil, fmt.Errorf("xmldom.ParseXML: unclosed elements: %d", len(stack)-1)
	}
	return root.Children, nil
}

func Marshal(nodes []*Node) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, n := range nodes {
		if err := n.Write(buf); err != nil {
			return nil, fmt.Errorf("xmldom.Marshal: %w", err)
		}
	}
	return buf.Bytes(), nil
}

func (n *Node) Write(w io.Writer) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "")
	if err := n.EncodeTo(enc); err != nil {
		return fmt.Errorf("xmldom.Write: %w", err)
	}
	if err := enc.Flush(); err != nil {
		return fmt.Errorf("xmldom.Write: %w", err)
	}
	return nil
}

func (n *Node) EncodeTo(enc *xml.Encoder) error {
	switch n.Kind {
	case ElementNode:
		if err := enc.EncodeToken(n.Start); err != nil {
			return err
		}
		for _, ch := range n.Children {
			if err := ch.EncodeTo(enc); err != nil {
				return err
			}
		}
		return enc.EncodeToken(n.End)
	case TextNode:
		return enc.EncodeToken(xml.CharData(n.Data))
	case CommentNode:
		return enc.EncodeToken(xml.Comment(n.Data))
	case DirectiveNode:
		return enc.EncodeToken(xml.Directive(n.Data))
	case ProcInstNode:
		return enc.EncodeToken(*n.PI)
	default:
		return fmt.Errorf("xmldom.Node.EncodeTo: unknown node kind: %v", n.Kind)
	}
}

func (n *Node) GetAttr(local string, space string) (string, bool) {
	for i := range n.Start.Attr {
		if n.Start.Attr[i].Name.Local == local && n.Start.Attr[i].Name.Space == space {
			return n.Start.Attr[i].Value, true
		}
	}
	return "", false
}

func (n *Node) SetAttr(local string, value string) {
	for i := range n.Start.Attr {
		if n.Start.Attr[i].Name.Local == local && n.Start.Attr[i].Name.Space == "" {
			n.Start.Attr[i].Value = value
			return
		}
	}
	n.Start.Attr = append(n.Start.Attr, xml.Attr{Name: xml.Name{Local: local}, Value: value})
}

func (n *Node) FloatAttr(local string) (float64, bool) {
	s, ok := n.GetAttr(local, "")
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func (n *Node) FirstChildElement(local string) *Node {
	for _, ch := range n.Children {
		if ch.Kind == ElementNode && ch.Start.Name.Local == local {
			return ch
		}
	}
	return nil
}

func (n *Node) HasChildElements() bool {
	for _, ch := range n.Children {
		if ch.Kind == ElementNode {
			return true
		}
	}
	return false
}

func (n *Node) IsBlankText() bool {
	return n.Kind == TextNode && strings.TrimSpace(string(n.Data)) == ""
}

func (n *Node) ChildElementIndent() string {
	for _, ch := range n.Children {
		if ch.Kind == ElementNode {
			break
		}
		if ch.IsBlankText() {
			return string(ch.Data)
		}
	}
	return "\n"
}

func (n *Node) Traverse(onEnter func(n *Node), onLeave func(n *Node)) {
	if onEnter != nil {
		onEnter(n)
	}
	for _, ch := range n.Children {
		ch.Traverse(onEnter, onLeave)
	}
	if onLeave != nil {
		onLeave(n)
	}
}

func (n *Node) Clone() *Node {
	cloned := &Node{
		Kind:     n.Kind,
		Start:    n.Start,
		End:      n.End,
		Children: make([]*Node, len(n.Children)),
		Data:     n.Data,
	}

	if n.Start.Attr != nil {
		cloned.Start.Attr = make([]xml.Attr, len(n.Start.Attr))
		copy(cloned.Start.Attr, n.Start.Attr)
	}
	for i, ch := range n.Children {
		cloned.Children[i] = ch.Clone()
	}
	return cloned
}

func (n *Node) AppendChildIndented(indent string, child *Node) {
	ind := NewText(indent)
	count := len(n.Children)
	if count > 0 && n.Children[count-1].IsBlankText() {
		tail := n.Children[count-1]
		newChildren := make([]*Node, 0, count+2)
		newChildren = append(newChildren, n.Children[:count-1]...)
		newChildren = append(newChildren, ind, child, tail)
		n.Children = newChildren
		return
	}
	n.Children = append(n.Children, ind, child)
}

func (n *Node) RemoveChildElements(pred func(*Node) bool) {
	var filtered []*Node
	blankOnly := true
	for _, ch := range n.Children {
		if ch.Kind == ElementNode && pred(ch) {
			if len(filtered) > 0 && filtered[len(filtered)-1].IsBlankText() {
				filtered = filtered[:len(filtered)-1]
			}
			continue
		}
		if !ch.IsBlankText() {
			blankOnly = false
		}
		filtered = append(filtered, ch)
	}

	n.Children = filtered
	if blankOnly {
		n.Children = nil
	}
}

func (n *Node) RewriteAttr(f func(node *Node, attr xml.Attr) xml.Attr) {
	n.Traverse(
		func(n *Node) {
			for i, attr := range n.Start.Attr {
				n.Start.Attr[i] = f(n, attr)
			}
		},
		nil,
	)
}
