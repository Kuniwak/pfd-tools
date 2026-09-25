package spreadsheet

import "testing"

func TestColumnLabel(t *testing.T) {
	cases := map[string]struct {
		n    int
		want string
	}{
		"1 -> A":      {n: 1, want: "A"},
		"2 -> B":      {n: 2, want: "B"},
		"26 -> Z":     {n: 26, want: "Z"},
		"27 -> AA":    {n: 27, want: "AA"},
		"28 -> AB":    {n: 28, want: "AB"},
		"52 -> AZ":    {n: 52, want: "AZ"},
		"53 -> BA":    {n: 53, want: "BA"},
		"702 -> ZZ":   {n: 702, want: "ZZ"},
		"703 -> AAA":  {n: 703, want: "AAA"},
		"0 -> empty":  {n: 0, want: ""},
		"-1 -> empty": {n: -1, want: ""},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := ColumnLabel(c.n)
			if got != c.want {
				t.Errorf("ColumnLabel(%d) = %q, want %q", c.n, got, c.want)
			}
		})
	}
}
