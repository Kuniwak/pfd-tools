package xmldom

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
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
	for i, ch := range n.Children {
		cloned.Children[i] = ch.Clone()
	}
	return cloned
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
