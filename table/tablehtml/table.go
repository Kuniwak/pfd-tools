package tablehtml

import (
	"fmt"
	"io"

	"golang.org/x/net/html"
)

type Table struct {
	Header []string
	Rows   [][]string
}

func NewTable(header []string, rows [][]string) (*Table, error) {
	col := len(header)
	for _, row := range rows {
		if len(row) != col {
			return nil, fmt.Errorf("tablehtml.NewTable: row length mismatch")
		}
	}
	return &Table{Header: header, Rows: rows}, nil
}

func RenderTable(w io.Writer, table *Table) error {
	tableNode := &html.Node{Type: html.ElementNode, Data: "table"}

	headerNode := &html.Node{Type: html.ElementNode, Data: "tr"}
	for _, header := range table.Header {
		headerCellNode := &html.Node{Type: html.ElementNode, Data: "th"}
		headerTextNode := &html.Node{Type: html.TextNode, Data: header}
		headerCellNode.AppendChild(headerTextNode)
		headerNode.AppendChild(headerCellNode)
	}
	tableNode.AppendChild(headerNode)

	for _, row := range table.Rows {
		rowNode := &html.Node{Type: html.ElementNode, Data: "tr"}

		for _, cell := range row {
			cellTextNode := &html.Node{Type: html.TextNode, Data: cell}

			cellNode := &html.Node{Type: html.ElementNode, Data: "td"}
			cellNode.AppendChild(cellTextNode)

			rowNode.AppendChild(cellNode)
		}

		tableNode.AppendChild(rowNode)
	}

	if err := html.Render(w, tableNode); err != nil {
		return fmt.Errorf("pfdhtml.RenderTable: %w", err)
	}
	return nil
}
