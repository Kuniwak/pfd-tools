package dot

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/sets"
)

type Digraph struct {
	Name            string
	GraphAttributes *sets.Set[*pairs.Pair[AttributeName, AttributeValue]]
	NodeAttributes  *sets.Set[*pairs.Pair[AttributeName, AttributeValue]]
	EdgeAttributes  *sets.Set[*pairs.Pair[AttributeName, AttributeValue]]
	Nodes           []*Node
	Edges           []*Edge
}

type NodeID string

type Node struct {
	ID         NodeID
	Attributes *sets.Set[*pairs.Pair[AttributeName, AttributeValue]]
}

func CompareAttribute(a, b *pairs.Pair[AttributeName, AttributeValue]) int {
	return pairs.Compare(CompareAttributeName, CompareAttributeValue)(a, b)
}

type AttributeName string

const (
	AttributeNameLabel     AttributeName = "label"
	AttributeNameShape     AttributeName = "shape"
	AttributeNameColor     AttributeName = "color"
	AttributeNameFillColor AttributeName = "fillcolor"
	AttributeNameCharset   AttributeName = "charset"
	AttributeNameStyle     AttributeName = "style"
	AttributeNameFontColor AttributeName = "fontcolor"
)

func (a AttributeName) String() string {
	return string(a)
}

func CompareAttributeName(a, b AttributeName) int {
	return strings.Compare(string(a), string(b))
}

type AttributeValueType byte

func CompareAttributeValueType(a, b AttributeValueType) int {
	return cmp.Compare(a, b)
}

const (
	AttributeValueTypeString AttributeValueType = iota
	AttributeValueTypeInt
	AttributeValueTypeFloat
	AttributeValueTypeBool
)

type AttributeValue struct {
	Type   AttributeValueType
	String string
	Int    int
	Float  float64
	Bool   bool
}

func CompareAttributeValue(a, b AttributeValue) int {
	c := cmp.Compare(a.Type, b.Type)
	if c != 0 {
		return c
	}
	switch a.Type {
	case AttributeValueTypeString:
		return strings.Compare(a.String, b.String)
	case AttributeValueTypeInt:
		return cmp.Compare(a.Int, b.Int)
	case AttributeValueTypeFloat:
		return cmp.Compare(a.Float, b.Float)
	case AttributeValueTypeBool:
		if a.Bool {
			if !b.Bool {
				return 1
			}
			return 0
		}
		return -1
	default:
		panic(fmt.Sprintf("pfddot.CompareAttributeValue: unknown attribute value type: %v", a.Type))
	}
}

type Edge struct {
	Source     NodeID
	Target     NodeID
	Attributes *sets.Set[*pairs.Pair[AttributeName, AttributeValue]]
}

func renderAttributeValue(w io.Writer, value AttributeValue, buf *bytes.Buffer) {
	switch value.Type {
	case AttributeValueTypeString:
		writeString(w, value.String, buf)
	case AttributeValueTypeInt:
		io.WriteString(w, strconv.Itoa(value.Int))
	case AttributeValueTypeFloat:
		io.WriteString(w, strconv.FormatFloat(value.Float, 'f', -1, 64))
	case AttributeValueTypeBool:
		io.WriteString(w, strconv.FormatBool(value.Bool))
	default:
		panic(fmt.Sprintf("pfddot.RenderAttributeValue: unknown attribute value type: %v", value.Type))
	}
}

func (d Digraph) Write(w io.Writer) error {
	buf := &bytes.Buffer{}
	io.WriteString(w, "digraph ")
	writeString(w, d.Name, buf)
	io.WriteString(w, " {\n  graph [")
	for i, pair := range d.GraphAttributes.Iter() {
		if i > 0 {
			io.WriteString(w, "; ")
		}
		name := pair.First
		value := pair.Second
		io.WriteString(w, name.String())
		io.WriteString(w, "=")
		renderAttributeValue(w, value, buf)
	}
	io.WriteString(w, "];\n  node [")
	for i, pair := range d.NodeAttributes.Iter() {
		if i > 0 {
			io.WriteString(w, "; ")
		}
		name := pair.First
		value := pair.Second
		io.WriteString(w, name.String())
		io.WriteString(w, "=")
		renderAttributeValue(w, value, buf)
	}
	io.WriteString(w, "];\n  edge [")
	for i, pair := range d.EdgeAttributes.Iter() {
		if i > 0 {
			io.WriteString(w, "; ")
		}
		name := pair.First
		value := pair.Second
		io.WriteString(w, name.String())
		io.WriteString(w, "=")
		renderAttributeValue(w, value, buf)
	}
	io.WriteString(w, "];\n")

	for _, node := range d.Nodes {
		io.WriteString(w, "  ")
		io.WriteString(w, string(node.ID))
		io.WriteString(w, " [")
		for _, pair := range node.Attributes.Iter() {
			name := pair.First
			value := pair.Second
			io.WriteString(w, name.String())
			io.WriteString(w, "=")
			renderAttributeValue(w, value, buf)
			io.WriteString(w, "; ")
		}
		io.WriteString(w, "];\n")
	}
	for _, edge := range d.Edges {
		io.WriteString(w, "  ")
		io.WriteString(w, string(edge.Source))
		io.WriteString(w, " -> ")
		io.WriteString(w, string(edge.Target))
		io.WriteString(w, " [")
		for _, pair := range edge.Attributes.Iter() {
			io.WriteString(w, pair.First.String())
			io.WriteString(w, "=")
			renderAttributeValue(w, pair.Second, buf)
			io.WriteString(w, "; ")
		}
		io.WriteString(w, "];\n")
	}

	io.WriteString(w, "}\n")
	return nil
}

func writeString(w io.Writer, s string, buf *bytes.Buffer) error {
	buf.Reset()
	e := json.NewEncoder(buf)
	e.SetEscapeHTML(false)
	if err := e.Encode(s); err != nil {
		return fmt.Errorf("fsmviz.writeString: %w", err)
	}
	bs := bytes.ReplaceAll(buf.Bytes(), []byte("\\n"), []byte("\\l"))
	w.Write(bs[0 : len(bs)-1]) // Remove trailing \n
	return nil
}
