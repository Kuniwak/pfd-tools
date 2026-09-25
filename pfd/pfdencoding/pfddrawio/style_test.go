package pfddrawio_test

import (
	"testing"

	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfddrawio"
	"github.com/google/go-cmp/cmp"
)

func TestStyleTokens(t *testing.T) {
	testCases := map[string]struct {
		style string
		want  []string
	}{
		"空":         {style: "", want: nil},
		"末尾のセミコロン":  {style: "html=1;", want: []string{"html=1"}},
		"複数":        {style: "edgeStyle=none;html=1;", want: []string{"edgeStyle=none", "html=1"}},
		"値なしのキー":    {style: "rounded=0;shadow;", want: []string{"rounded=0", "shadow"}},
		"末尾セミコロンなし": {style: "html=1", want: []string{"html=1"}},
		"空トークンは捨てる": {style: ";html=1;;dashed=1;", want: []string{"html=1", "dashed=1"}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.StyleTokens(tc.style)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("StyleTokens(%q) mismatch (-want +got):\n%s", tc.style, diff)
			}
		})
	}
}

func TestStyleTokenName(t *testing.T) {
	testCases := map[string]struct {
		token string
		want  string
	}{
		"キーと値":     {token: "entryX=0", want: "entryX"},
		"値なし":      {token: "html", want: "html"},
		"値に = を含む": {token: "label=a=b", want: "label"},
		"空":        {token: "", want: ""},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := pfddrawio.StyleTokenName(tc.token); got != tc.want {
				t.Errorf("StyleTokenName(%q) = %q, want %q", tc.token, got, tc.want)
			}
		})
	}
}

func TestWithStyleToken(t *testing.T) {
	testCases := map[string]struct {
		tokens []string
		name   string
		value  string
		want   []string
	}{
		"既存のキーはその位置で置換する": {
			tokens: []string{"edgeStyle=orthogonalEdgeStyle", "html=1"},
			name:   "edgeStyle",
			value:  "none",
			want:   []string{"edgeStyle=none", "html=1"},
		},
		"無いキーは末尾に足す": {
			tokens: []string{"html=1"},
			name:   "jumpStyle",
			value:  "gap",
			want:   []string{"html=1", "jumpStyle=gap"},
		},
		"空のトークン列": {
			tokens: nil,
			name:   "edgeStyle",
			value:  "none",
			want:   []string{"edgeStyle=none"},
		},
		"重複するキーはすべて置換して 1 つに畳む": {
			tokens: []string{"jumpStyle=arc", "html=1", "jumpStyle=none"},
			name:   "jumpStyle",
			value:  "gap",
			want:   []string{"jumpStyle=gap", "html=1"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.WithStyleToken(tc.tokens, tc.name, tc.value)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WithStyleToken mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWithoutStyleTokens(t *testing.T) {
	testCases := map[string]struct {
		tokens []string
		names  []string
		want   []string
	}{
		"指定したキーを消す": {
			tokens: []string{"entryX=0", "html=1", "entryY=0.5"},
			names:  []string{"entryX", "entryY"},
			want:   []string{"html=1"},
		},
		"無いキーは無視する": {
			tokens: []string{"html=1"},
			names:  []string{"exitX"},
			want:   []string{"html=1"},
		},
		"すべて消える": {
			tokens: []string{"entryX=0"},
			names:  []string{"entryX"},
			want:   nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.WithoutStyleTokens(tc.tokens, tc.names...)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WithoutStyleTokens mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatStyle(t *testing.T) {
	testCases := map[string]struct {
		tokens []string
		want   string
	}{
		"空は空文字列":   {tokens: nil, want: ""},
		"末尾にセミコロン": {tokens: []string{"html=1"}, want: "html=1;"},
		"複数":       {tokens: []string{"edgeStyle=none", "html=1"}, want: "edgeStyle=none;html=1;"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if got := pfddrawio.FormatStyle(tc.tokens); got != tc.want {
				t.Errorf("FormatStyle(%v) = %q, want %q", tc.tokens, got, tc.want)
			}
		})
	}
}

func TestStraightEdgeStyle(t *testing.T) {
	testCases := map[string]struct {
		style string
		want  string
	}{
		"直交ルーティングを直線にする": {
			style: "edgeStyle=orthogonalEdgeStyle;html=1;jumpStyle=gap;",
			want:  "edgeStyle=none;html=1;jumpStyle=gap;",
		},
		"接続点の固定を外す": {
			style: "edgeStyle=orthogonalEdgeStyle;shape=connector;rounded=1;jumpStyle=gap;html=1;entryX=0;entryY=0.5;entryDx=0;entryDy=0;strokeColor=default;endArrow=classic;",
			want:  "edgeStyle=none;shape=connector;rounded=1;jumpStyle=gap;html=1;strokeColor=default;endArrow=classic;",
		},
		"exit 側の接続点の固定も外す": {
			style: "edgeStyle=orthogonalEdgeStyle;html=1;exitX=1;exitY=0.5;exitDx=0;exitDy=0;",
			want:  "edgeStyle=none;html=1;jumpStyle=gap;",
		},
		"接続点の周辺スナップ指定も外す": {
			style: "edgeStyle=orthogonalEdgeStyle;html=1;entryX=0;entryY=0.5;entryPerimeter=0;exitPerimeter=0;",
			want:  "edgeStyle=none;html=1;jumpStyle=gap;",
		},
		"jumpStyle が無ければ足す": {
			style: "edgeStyle=none;html=1;",
			want:  "edgeStyle=none;html=1;jumpStyle=gap;",
		},
		"jumpStyle が別の値なら置き換える": {
			style: "edgeStyle=orthogonalEdgeStyle;jumpStyle=arc;html=1;",
			want:  "edgeStyle=none;jumpStyle=gap;html=1;",
		},
		"破線などその他のキーと並びは保つ": {
			style: "edgeStyle=orthogonalEdgeStyle;shape=connector;rounded=1;jumpStyle=gap;html=1;dashed=1;strokeColor=default;endArrow=classic;",
			want:  "edgeStyle=none;shape=connector;rounded=1;jumpStyle=gap;html=1;dashed=1;strokeColor=default;endArrow=classic;",
		},
		"すでに直線ならそのまま": {
			style: "edgeStyle=none;html=1;jumpStyle=gap;",
			want:  "edgeStyle=none;html=1;jumpStyle=gap;",
		},
		"空の style": {
			style: "",
			want:  "edgeStyle=none;jumpStyle=gap;",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := pfddrawio.StraightEdgeStyle(tc.style)
			if got != tc.want {
				t.Errorf("StraightEdgeStyle(%q) =\n %q\nwant\n %q", tc.style, got, tc.want)
			}

			if again := pfddrawio.StraightEdgeStyle(got); again != got {
				t.Errorf("StraightEdgeStyle is not idempotent: %q -> %q", got, again)
			}
		})
	}
}

func TestParseStyle(t *testing.T) {
	testCases := map[string]struct {
		style   string
		want    pfddrawio.StyleMap
		wantErr bool
	}{
		"空":          {style: "", want: pfddrawio.StyleMap{}},
		"キーと値":       {style: "rounded=0;html=1;", want: pfddrawio.StyleMap{"rounded": "0", "html": "1"}},
		"値なしのキー":     {style: "ellipse;html=1;", want: pfddrawio.StyleMap{"ellipse": "", "html": "1"}},
		"重複するキーは後勝ち": {style: "html=0;html=1;", want: pfddrawio.StyleMap{"html": "1"}},
		"値に = を含むトークンはエラー": {style: "label=a=b;", wantErr: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := pfddrawio.ParseStyle(tc.style)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseStyle(%q) err = nil, want an error", tc.style)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseStyle(%q): %v", tc.style, err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ParseStyle(%q) mismatch (-want +got):\n%s", tc.style, diff)
			}
		})
	}
}
