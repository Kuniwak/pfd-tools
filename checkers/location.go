package checkers

import (
	"io"
	"strings"
)

type Location interface {
	Write(w io.Writer) error
}

func CompareLocation(a, b Location, sb *strings.Builder) int {
	sb.Reset()
	if err := a.Write(sb); err != nil {
		panic(err)
	}
	as := sb.String()
	sb.Reset()
	if err := b.Write(sb); err != nil {
		panic(err)
	}
	bs := sb.String()
	return strings.Compare(as, bs)
}
