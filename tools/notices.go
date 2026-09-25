package tools

import "strings"

type NoticeModule struct {
	Path    string
	Version string
}

func ParseModuleList(output string) []NoticeModule {
	var modules []NoticeModule
	seen := map[NoticeModule]bool{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		m := NoticeModule{Path: fields[0], Version: fields[1]}
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}
	return modules
}

func MissingNoticeModules(modules []NoticeModule, notices string) []NoticeModule {
	var missing []NoticeModule
	for _, m := range modules {
		if !strings.Contains(notices, "== "+m.Path+" ==\nVersion: "+m.Version+"\n") {
			missing = append(missing, m)
		}
	}
	return missing
}
