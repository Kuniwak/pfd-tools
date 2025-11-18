package allcheckers

import (
	"fmt"

	"github.com/Kuniwak/pfd-tools/checkers"
	"github.com/Kuniwak/pfd-tools/locale"
)

func Message(id checkers.ProblemID, l locale.Locale) string {
	switch l {
	case locale.LocaleJa:
		return JapaneseMessage(id)
	case locale.LocaleEn:
		return EnglishMessage(id)
	default:
		panic(fmt.Sprintf("allcheckers.Message: unknown locale: %q", l))
	}
}

func EnglishMessage(id checkers.ProblemID) string {
	switch id {
	case "no-desc":
		return "Please add a concise description."
	case "consistent-desc":
		return "elements with the same ID should have the same description."
	case "in-field":
		return "elements appearing on both ends of an edge or feedback edge should be included in the element set."
	case "ex-input":
		return "At least one input deliverable is required."
	case "ex-output":
		return "At least one output deliverable is required."
	case "no-d2d":
		return "A deliverable should not be directly connected to another deliverable."
	case "no-p2p":
		return "A process should not be directly connected to another process."
	case "no-p2d-fb":
		return "A feedback edge should connect a deliverable to a process."
	case "single-src":
		return "A deliverable should be output from only one process. This includes output through feedback edges."
	case "acyclic-except-fb":
		return "If the feedback edge is removed, there is a cycle in the graph."
	case "cyclic-ex1-fb":
		return "A feedback loop should contain only one feedback edge."
	case "weak-conn":
		return "There is a final deliverable that is not reachable from any initial deliverable."
	case "finite":
		return "The process set, deliverable set, and edge set should allcheckers be finite sets."
	case "disj-or-psubset-comp":
		return "Different composite processes should be disjoint or one should be a proper subset of the other."
	case "consistent-input-comp":
		return "The input deliverable set of a composite process does not match the input of the atomic processes it contains."
	case "consistent-output-comp":
		return "The output deliverable set of a composite process does not match the output of the atomic processes it contains."
	case "valid-available-time":
		return "The available time should be a non-negative 64bit float."
	case "valid-init-volume":
		return "The initial volume should be a non-negative number."
	case "malformed-max-revision":
		return "The max revision should be a 1 or greater integer."
	case "malformed-resources-set-notation":
		return "The resources set should be a ;-separated string. Each entry should be <resource ID>,<resource ID>,...:<non-negative floating point number>."
	case "empty-resources-set":
		return "The resources set should not be empty."
	case "zero-consumed-volume":
		return "The consumed volume should not be zero."
	case "no-zero-volume-fb":
		return "The initial volume of an atomic process that is the destination of a feedback edge should be zero."
	case "missing-r-table":
		return "The resource ID is missing from the resource table."
	case "extra-r-table":
		return "The resource ID is extra from the resource table."
	case "missing-ap-table":
		return "The atomic process ID is missing from the atomic process table."
	case "extra-ap-table":
		return "The atomic process ID is extra from the atomic process table."
	case "missing-cp-table":
		return "The composite process ID is missing from the composite process table."
	case "extra-cp-table":
		return "The composite process ID is extra from the composite process table."
	case "missing-d-table":
		return "The deliverable ID is missing from the deliverable table."
	case "extra-d-table":
		return "The deliverable ID is extra from the deliverable table."
	case "malformed-precondition":
		return "The precondition has an syntax error."
	case "precondition-not-feedback":
		return "The precondition should be a feedback edge."
	case "malformed-g-table":
		return "The group ID has a syntax error."
	case "missing-g-table":
		return "The group ID is missing from the group table."
	case "extra-g-table":
		return "The group ID is extra from the group table."
	case "malformed-m-table":
		return "The milestone ID has a syntax error."
	case "malformed-m-table-successors":
		return "The successors has a syntax error."
	case "missing-m-table":
		return "The milestone ID is missing from the milestone table."
	case "extra-m-table":
		return "The milestone ID is extra from the milestone table."
	}
	panic(fmt.Sprintf("unknown problem ID: %q", id))
}

func JapaneseMessage(id checkers.ProblemID) string {
	switch id {
	case "no-desc":
		return "端的な説明を追加してください。"
	case "consistent-desc":
		return "同じIDの要素（複製表示）は説明が一致すべきです。"
	case "in-field":
		return "辺またはフィードバック辺の両端にあらわれる要素は要素集合に含まれていなければなりません。"
	case "ex-input":
		return "1つ以上の入力成果物が必要です。"
	case "ex-output":
		return "1つ以上の出力成果物が必要です。"
	case "no-d2d":
		return "成果物と成果物を直接結んではいけません。"
	case "no-p2p":
		return "プロセスとプロセスを直接結んではいけません。"
	case "no-p2d-fb":
		return "フィードバック辺は成果物からプロセスを結ぶ必要があります。"
	case "single-src":
		return "成果物が複数のプロセスから出力されています。成果物はただ1つのプロセスから出力されるべきです。"
	case "acyclic-except-fb":
		return "フィードバック辺を取り除くとグラフに循環路があります。"
	case "cyclic-ex1-fb":
		return "フィードバックループにはただ1つのフィードバック辺が含まれる必要があります。"
	case "weak-conn":
		return "初期成果物から到達できない最終成果物があります。"
	case "finite":
		return "プロセス集合、成果物集合、辺集合はいずれも有限集合でなければなりません。"
	case "disj-or-psubset-comp":
		return "異なる複合プロセスは互いに素であるか一方が一方の真部分集合でなければなりません。"
	case "consistent-input-comp":
		return "複合プロセスの入力成果物集合が内包する原子プロセスの入力と整合しません。"
	case "consistent-output-comp":
		return "複合プロセスの出力成果物集合が内包する原子プロセスの出力と整合しません。"
	case "valid-available-time":
		return "利用可能時間は非負浮動小数点数でなければなりません。"
	case "valid-init-volume":
		return "初期作業量は非負数でなければなりません。"
	case "malformed-max-revision":
		return "最大版数は各成果物について1以上の整数でなければなりません。"
	case "malformed-resources-set-notation":
		return "資源集合は;区切りの1行のみのCSVでなければなりません。"
	case "empty-resources-set":
		return "資源集合は空でなければなりません。"
	case "zero-consumed-volume":
		return "消費作業量は0でなければなりません。"
	case "no-zero-volume-fb":
		return "フィードバック辺の先の原子プロセスの初期作業量は0でなければなりません。"
	case "missing-r-table":
		return "資源IDが資源表にありません。"
	case "extra-r-table":
		return "資源IDが資源表に余分です。"
	case "missing-ap-table":
		return "原子プロセスIDが原子プロセス表にありません。"
	case "extra-ap-table":
		return "原子プロセスIDが原子プロセス表に余分です。"
	case "missing-cp-table":
		return "複合プロセスIDが複合プロセス表にありません。"
	case "extra-cp-table":
		return "複合プロセスIDが複合プロセス表に余分です。"
	case "missing-d-table":
		return "成果物IDが成果物表にありません。"
	case "extra-d-table":
		return "成果物IDが成果物表に余分です。"
	case "malformed-precondition":
		return "開始条件に構文エラーがあります。"
	case "precondition-not-feedback":
		return "開始条件はフィードバック辺でなければなりません。"
	case "malformed-g-table":
		return "グループIDに構文エラーがあります。"
	case "malformed-m-table-successors":
		return "後続マイルストーンに構文エラーがあります。"
	case "missing-m-table":
		return "マイルストーンIDがマイルストーン表にありません。"
	case "extra-m-table":
		return "マイルストーンIDがマイルストーン表に余分です。"
	case "missing-g-table":
		return "グループIDがグループ表にありません。"
	case "extra-g-table":
		return "グループIDがグループ表に余分です。"
	default:
		panic(fmt.Sprintf("unknown problem ID: %q", id))
	}
}
