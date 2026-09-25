package pfdticket

import (
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/sets"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func adTable(rows ...*pfd.AtomicDeliverableRow) *pfd.AtomicDeliverableTable {
	return &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{"利用可能時刻", "最大版", "フォーマット", "レビュー基準", "レビューア", "URL"},
		Rows:         rows,
	}
}

func adRow(id, desc, format, reviewCriteria, reviewer, url string) *pfd.AtomicDeliverableRow {
	return &pfd.AtomicDeliverableRow{
		ID:          pfd.AtomicDeliverableID(id),
		Description: desc,
		ExtraCells:  []string{"", "", format, reviewCriteria, reviewer, url},
	}
}

func node(id, desc string, t pfd.NodeType) *pfd.Node {
	return &pfd.Node{ID: pfd.NodeID(id), Description: desc, Type: t}
}

func TestGenerateTickets(t *testing.T) {
	testCases := map[string]struct {
		pfd  *pfd.PFD
		ad   *pfd.AtomicDeliverableTable
		want []Ticket
	}{
		"single process with review criteria and reviewer": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "レビュー指摘が全てクローズ", "@alice", ""),
				adRow("D2", "設計書", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 品質基準: レビュー指摘が全てクローズ
  - レビューア: @alice
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"output url already filled uses url instead of placeholder": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "レビューOK", "@alice", "https://example.com/pr/1"),
				adRow("D2", "設計書", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 品質基準: レビューOK
  - レビューア: @alice
  - 成果物リンク: https://example.com/pr/1`,
				},
			},
		},

		"empty review columns are omitted but link line is always present": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "", "", ""),
				adRow("D2", "設計書", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"input format and url are shown in parentheses": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "", "", ""),
				adRow("D2", "設計書", "Google Docs", "", "", "https://example.com/design"),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（フォーマット: Google Docs／所在: https://example.com/design）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"multiple inputs and outputs sorted by id": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "設計書", pfd.NodeTypeAtomicDeliverable),
					node("D2", "要件定義", pfd.NodeTypeAtomicDeliverable),
					node("D3", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D4", "テスト", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D1", Target: "P1"},
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D3"},
					&pfd.Edge{Source: "P1", Target: "D4"},
				),
			},
			ad: adTable(
				adRow("D1", "設計書", "", "", "", ""),
				adRow("D2", "要件定義", "", "", "", ""),
				adRow("D3", "実装コード", "", "レビューOK", "@alice", ""),
				adRow("D4", "テスト", "", "全て緑", "@bob", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D1: 設計書（所在: 成果物所在不明）
- D2: 要件定義（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D3: 実装コード
  - 品質基準: レビューOK
  - レビューア: @alice
  - 成果物リンク: （完了時に URL を記入）
- D4: テスト
  - 品質基準: 全て緑
  - レビューア: @bob
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"context diagram P0 is excluded": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P0", "全体", pfd.NodeTypeAtomicProcess),
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "", "", ""),
				adRow("D2", "設計書", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"feedback edges are excluded from lists": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
					node("D3", "レビュー指摘", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
					&pfd.Edge{Source: "D3", Target: "P1", IsFeedback: true},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "", "", ""),
				adRow("D2", "設計書", "", "", "", ""),
				adRow("D3", "レビュー指摘", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"without ad table falls back to node descriptions": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: nil,
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"process without inputs or outputs shows none marker": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "調査", pfd.NodeTypeAtomicProcess),
				),
				Edges: sets.New((*pfd.Edge).Compare),
			},
			ad: adTable(),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 調査",
					Description: `# 入力成果物の一覧

（なし）

# 出力成果物の一覧と品質基準とレビューア

（なし）`,
				},
			},
		},
		"input produced by another process links to producing process ticket": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "設計", pfd.NodeTypeAtomicProcess),
					node("P2", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "設計書", pfd.NodeTypeAtomicDeliverable),
					node("D2", "実装コード", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "P1", Target: "D1"},
					&pfd.Edge{Source: "D1", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D2"},
				),
			},
			ad: adTable(
				adRow("D1", "設計書", "", "", "", ""),
				adRow("D2", "実装コード", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 設計",
					Description: `# 入力成果物の一覧

（なし）

# 出力成果物の一覧と品質基準とレビューア

- D1: 設計書
  - 成果物リンク: （完了時に URL を記入）`,
				},
				{
					ID:      "P2",
					Summary: "P2: 実装",
					Description: `# 入力成果物の一覧

- D1: 設計書（所在: P1 のチケット）

# 出力成果物の一覧と品質基準とレビューア

- D2: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"input produced via feedback edge shows producing process ticket": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "レビュー", pfd.NodeTypeAtomicProcess),
					node("P2", "修正", pfd.NodeTypeAtomicProcess),
					node("D1", "指摘", pfd.NodeTypeAtomicDeliverable),
					node("D2", "修正版", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "P1", Target: "D1", IsFeedback: true},
					&pfd.Edge{Source: "D1", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D2"},
				),
			},
			ad: adTable(
				adRow("D1", "指摘", "", "", "", ""),
				adRow("D2", "修正版", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: レビュー",
					Description: `# 入力成果物の一覧

（なし）

# 出力成果物の一覧と品質基準とレビューア

（なし）`,
				},
				{
					ID:      "P2",
					Summary: "P2: 修正",
					Description: `# 入力成果物の一覧

- D1: 指摘（所在: P1 のチケット）

# 出力成果物の一覧と品質基準とレビューア

- D2: 修正版
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"initial input with url only shows url as location": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
					node("D2", "設計書", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTable(
				adRow("D1", "実装コード", "", "", "", ""),
				adRow("D2", "設計書", "", "", "", "https://example.com/design"),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 実装",
					Description: `# 入力成果物の一覧

- D2: 設計書（所在: https://example.com/design）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},

		"multiple input location forms in one process": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "設計", pfd.NodeTypeAtomicProcess),
					node("P2", "実装", pfd.NodeTypeAtomicProcess),
					node("D1", "設計書", pfd.NodeTypeAtomicDeliverable),
					node("D2", "要件", pfd.NodeTypeAtomicDeliverable),
					node("D3", "メモ", pfd.NodeTypeAtomicDeliverable),
					node("D4", "コード", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "P1", Target: "D1"},
					&pfd.Edge{Source: "D1", Target: "P2"},
					&pfd.Edge{Source: "D2", Target: "P2"},
					&pfd.Edge{Source: "D3", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D4"},
				),
			},
			ad: adTable(
				adRow("D1", "設計書", "", "", "", ""),
				adRow("D2", "要件", "", "", "", "https://example.com/req"),
				adRow("D3", "メモ", "", "", "", ""),
				adRow("D4", "コード", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: 設計",
					Description: `# 入力成果物の一覧

（なし）

# 出力成果物の一覧と品質基準とレビューア

- D1: 設計書
  - 成果物リンク: （完了時に URL を記入）`,
				},
				{
					ID:      "P2",
					Summary: "P2: 実装",
					Description: `# 入力成果物の一覧

- D1: 設計書（所在: P1 のチケット）
- D2: 要件（所在: https://example.com/req）
- D3: メモ（所在: 成果物所在不明）

# 出力成果物の一覧と品質基準とレビューア

- D4: コード
  - 成果物リンク: （完了時に URL を記入）`,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := GenerateTickets(tc.pfd, tc.ad, DefaultConfig(locale.LocaleJa), testLogger())
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GenerateTickets() mismatch\n--- got ---\n%s\n--- want ---\n%s", formatTickets(got), formatTickets(tc.want))
			}
		})
	}
}

func adTableEn(rows ...*pfd.AtomicDeliverableRow) *pfd.AtomicDeliverableTable {
	return &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{"Available Time", "Max Revision", "Format", "Review Criteria", "Reviewer", "URL"},
		Rows:         rows,
	}
}

func TestGenerateTicketsLocaleEn(t *testing.T) {
	testCases := map[string]struct {
		pfd  *pfd.PFD
		ad   *pfd.AtomicDeliverableTable
		want []Ticket
	}{
		"initial input shows unknown location": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "Implement", pfd.NodeTypeAtomicProcess),
					node("D1", "Code", pfd.NodeTypeAtomicDeliverable),
					node("D2", "Design", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "D2", Target: "P1"},
					&pfd.Edge{Source: "P1", Target: "D1"},
				),
			},
			ad: adTableEn(
				adRow("D1", "Code", "", "All review comments closed", "@alice", ""),
				adRow("D2", "Design", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: Implement",
					Description: `# Input Deliverables

- D2: Design（Location: deliverable location unknown）

# Output Deliverables with Quality Criteria and Reviewers

- D1: Code
  - Quality Criteria: All review comments closed
  - Reviewer: @alice
  - Deliverable Link: (fill in the URL when done)`,
				},
			},
		},

		"input produced by another process links to producing process ticket": {
			pfd: &pfd.PFD{
				Nodes: sets.New((*pfd.Node).Compare,
					node("P1", "Design", pfd.NodeTypeAtomicProcess),
					node("P2", "Implement", pfd.NodeTypeAtomicProcess),
					node("D1", "Design Doc", pfd.NodeTypeAtomicDeliverable),
					node("D2", "Code", pfd.NodeTypeAtomicDeliverable),
				),
				Edges: sets.New((*pfd.Edge).Compare,
					&pfd.Edge{Source: "P1", Target: "D1"},
					&pfd.Edge{Source: "D1", Target: "P2"},
					&pfd.Edge{Source: "P2", Target: "D2"},
				),
			},
			ad: adTableEn(
				adRow("D1", "Design Doc", "", "", "", ""),
				adRow("D2", "Code", "", "", "", ""),
			),
			want: []Ticket{
				{
					ID:      "P1",
					Summary: "P1: Design",
					Description: `# Input Deliverables

(none)

# Output Deliverables with Quality Criteria and Reviewers

- D1: Design Doc
  - Deliverable Link: (fill in the URL when done)`,
				},
				{
					ID:      "P2",
					Summary: "P2: Implement",
					Description: `# Input Deliverables

- D1: Design Doc（Location: P1 ticket）

# Output Deliverables with Quality Criteria and Reviewers

- D2: Code
  - Deliverable Link: (fill in the URL when done)`,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := GenerateTickets(tc.pfd, tc.ad, DefaultConfig(locale.LocaleEn), testLogger())
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GenerateTickets(en) mismatch\n--- got ---\n%s\n--- want ---\n%s", formatTickets(got), formatTickets(tc.want))
			}
		})
	}
}

func TestGenerateTicketsCustomColumns(t *testing.T) {
	p := &pfd.PFD{
		Nodes: sets.New((*pfd.Node).Compare,
			node("P1", "実装", pfd.NodeTypeAtomicProcess),
			node("D1", "実装コード", pfd.NodeTypeAtomicDeliverable),
		),
		Edges: sets.New((*pfd.Edge).Compare,
			&pfd.Edge{Source: "P1", Target: "D1"},
		),
	}

	ad := &pfd.AtomicDeliverableTable{
		ExtraHeaders: []string{"品質", "担当", "リンク"},
		Rows: []*pfd.AtomicDeliverableRow{
			{ID: "D1", Description: "実装コード", ExtraCells: []string{"レビューOK", "@alice", "https://example.com/pr/1"}},
		},
	}
	config := DefaultConfig(locale.LocaleJa)
	config.ReviewCriteriaColumn = "品質"
	config.ReviewerColumn = "担当"
	config.URLColumn = "リンク"

	got := GenerateTickets(p, ad, config, testLogger())
	want := []Ticket{
		{
			ID:      "P1",
			Summary: "P1: 実装",
			Description: `# 入力成果物の一覧

（なし）

# 出力成果物の一覧と品質基準とレビューア

- D1: 実装コード
  - 品質基準: レビューOK
  - レビューア: @alice
  - 成果物リンク: https://example.com/pr/1`,
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GenerateTickets(custom columns) mismatch\n--- got ---\n%s\n--- want ---\n%s", formatTickets(got), formatTickets(want))
	}
}

func formatTickets(tickets []Ticket) string {
	s := ""
	for _, ticket := range tickets {
		s += "ID=" + ticket.ID + "\nSummary=" + ticket.Summary + "\nDescription:\n" + ticket.Description + "\n===\n"
	}
	return s
}
