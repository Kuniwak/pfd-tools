package tools

import (
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestParseModuleList(t *testing.T) {
	testCases := map[string]struct {
		output   string
		expected []NoticeModule
	}{
		"empty": {
			output:   "",
			expected: nil,
		},
		"skips blank lines and duplicates": {
			output: "golang.org/x/net v0.55.0\n\ngolang.org/x/sync v0.17.0\ngolang.org/x/net v0.55.0\n",
			expected: []NoticeModule{
				{Path: "golang.org/x/net", Version: "v0.55.0"},
				{Path: "golang.org/x/sync", Version: "v0.17.0"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := ParseModuleList(tc.output)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("ParseModuleList() = %v, want %v", actual, tc.expected)
			}
		})
	}
}

func TestMissingNoticeModules(t *testing.T) {
	net := NoticeModule{Path: "golang.org/x/net", Version: "v0.55.0"}
	sync := NoticeModule{Path: "golang.org/x/sync", Version: "v0.17.0"}

	testCases := map[string]struct {
		modules  []NoticeModule
		notices  string
		expected []NoticeModule
	}{
		"no modules": {
			modules:  nil,
			notices:  "",
			expected: nil,
		},
		"all listed": {
			modules:  []NoticeModule{net, sync},
			notices:  "== golang.org/x/net ==\nVersion: v0.55.0\n...\n== golang.org/x/sync ==\nVersion: v0.17.0\n...\n",
			expected: nil,
		},
		"one missing": {
			modules:  []NoticeModule{net, sync},
			notices:  "== golang.org/x/net ==\nVersion: v0.55.0\n...\n",
			expected: []NoticeModule{sync},
		},
		"version mismatch": {
			modules:  []NoticeModule{net},
			notices:  "== golang.org/x/net ==\nVersion: v0.54.0\n...\n",
			expected: []NoticeModule{net},
		},
		"prefix does not count": {
			modules:  []NoticeModule{net},
			notices:  "== golang.org/x/network ==\nVersion: v0.55.0\n",
			expected: []NoticeModule{net},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := MissingNoticeModules(tc.modules, tc.notices)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("MissingNoticeModules() = %v, want %v", actual, tc.expected)
			}
		})
	}
}

func distributedModules(t *testing.T) []NoticeModule {
	t.Helper()
	cmd := exec.Command("go", "list", "-deps", "-f", "{{with .Module}}{{if not .Main}}{{.Path}} {{.Version}}{{end}}{{end}}", "./tools/...")
	cmd.Dir = ".."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	return ParseModuleList(string(out))
}

func TestThirdPartyNoticesCoverDistributedModules(t *testing.T) {
	modules := distributedModules(t)
	notices, err := os.ReadFile("../THIRD_PARTY_NOTICES")
	if err != nil {
		t.Fatalf("read THIRD_PARTY_NOTICES: %v", err)
	}

	missing := MissingNoticeModules(modules, string(notices))

	if len(missing) > 0 {
		t.Errorf("THIRD_PARTY_NOTICES に次のモジュールの表示がないか、バージョンが食い違っています: %v", missing)
	}
}

func TestDistributedModulesExcludeTestOnlyModules(t *testing.T) {
	testOnly := map[string]bool{"pgregory.net/rapid": true, "github.com/google/go-cmp": true}

	modules := distributedModules(t)

	for _, m := range modules {
		if testOnly[m.Path] {
			t.Errorf("テスト専用のモジュールが配布バイナリに入っています: %s", m.Path)
		}
	}
}
