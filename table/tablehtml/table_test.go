package tablehtml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewTable(t *testing.T) {
	table, err := NewTable([]string{"Column1", "Column2"}, [][]string{{"1-1", "1-2"}, {"2-1", "2-2"}})
	if err != nil {
		t.Fatal(err)
	}

	want := strings.ReplaceAll(strings.ReplaceAll(`
	<table>
		<tr>
			<th>Column1</th>
			<th>Column2</th>
		</tr>
		<tr>
			<td>1-1</td>
			<td>1-2</td>
		</tr>
		<tr>
			<td>2-1</td>
			<td>2-2</td>
		</tr>
	</table>
	`, "\n", ""), "\t", "")

	buf := bytes.NewBuffer(nil)
	if err := RenderTable(buf, table); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if got != want {
		t.Error(cmp.Diff(want, got))
	}
}
