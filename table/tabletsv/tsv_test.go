package tabletsv

import "testing"

func TestRequireColumns(t *testing.T) {
	testCases := map[string]struct {
		Header  []string
		Min     int
		WantErr bool
	}{
		"fewer columns than required": {
			Header:  []string{"ID"},
			Min:     2,
			WantErr: true,
		},
		"no columns at all": {
			Header:  []string{},
			Min:     2,
			WantErr: true,
		},
		"exactly the required columns": {
			Header:  []string{"ID", "Description"},
			Min:     2,
			WantErr: false,
		},
		"more columns than required": {
			Header:  []string{"ID", "Description", "Note"},
			Min:     2,
			WantErr: false,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := RequireColumns(tc.Header, tc.Min)
			if tc.WantErr && err == nil {
				t.Fatalf("expected an error for header %v, got nil", tc.Header)
			}
			if !tc.WantErr && err != nil {
				t.Fatalf("expected no error for header %v, got %v", tc.Header, err)
			}
		})
	}
}

func TestRequireUniqueFirstColumn(t *testing.T) {
	testCases := map[string]struct {
		Rows    [][]string
		WantErr bool
	}{
		"no rows": {
			Rows:    [][]string{},
			WantErr: false,
		},
		"distinct IDs": {
			Rows:    [][]string{{"P1", "実装"}, {"P2", "検査"}},
			WantErr: false,
		},

		"the same ID twice": {
			Rows:    [][]string{{"P1", "実装"}, {"P1", "別の実装"}},
			WantErr: true,
		},
		"the same ID with the same content": {
			Rows:    [][]string{{"P1", "実装"}, {"P1", "実装"}},
			WantErr: true,
		},

		"a row without the ID column": {
			Rows:    [][]string{{}},
			WantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := RequireUniqueFirstColumn(tc.Rows)
			if tc.WantErr && err == nil {
				t.Fatalf("expected an error for rows %v, got nil", tc.Rows)
			}
			if !tc.WantErr && err != nil {
				t.Fatalf("expected no error for rows %v, got %v", tc.Rows, err)
			}
		})
	}
}
