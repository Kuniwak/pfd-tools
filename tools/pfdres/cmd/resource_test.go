package cmd

import "testing"

func TestNeededResourceSets(t *testing.T) {
	cases := map[string]struct {
		headCount int
		want      string
	}{
		"1":  {headCount: 1, want: "A:1"},
		"2":  {headCount: 2, want: "A:1;B:1"},
		"3":  {headCount: 3, want: "A:1;B:1;C:1"},
		"27": {headCount: 27, want: "A:1;B:1;C:1;D:1;E:1;F:1;G:1;H:1;I:1;J:1;K:1;L:1;M:1;N:1;O:1;P:1;Q:1;R:1;S:1;T:1;U:1;V:1;W:1;X:1;Y:1;Z:1;AA:1"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := NeededResourceSets(c.headCount)
			if got != c.want {
				t.Errorf("NeededResourceSets(%d) = %q, want %q", c.headCount, got, c.want)
			}
		})
	}
}
