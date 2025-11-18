package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/locale"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/pfdencoding/pfdfmt"
	"github.com/Kuniwak/pfd-tools/pfd/pfdtable/encoding/pfdtsv"
	"github.com/Kuniwak/pfd-tools/version"
)

func MainCommandByArgs(args []string, inout *cli.ProcInout) int {
	opts, err := ParseOptions(args, inout)
	if err != nil {
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	if err := MainCommandByOptions(opts, inout); err != nil {
		if errors.Is(err, ErrHasDiff) {
			return 1
		}
		fmt.Fprintln(inout.Stderr, err.Error())
		return 1
	}
	return 0
}

var ErrHasDiff = errors.New("cmd.MainCommandByOptions: diff is not empty")

func MainCommandByOptions(opts *Options, inout *cli.ProcInout) error {
	if opts.CommonOptions.Help {
		return nil
	}

	if opts.CommonOptions.Version {
		fmt.Fprintln(inout.Stdout, version.Version)
		return nil
	}

	if opts.Prompt {
		r, w := io.Pipe()
		go func() {
			defer w.Close()
			diff, err := buildDiff(opts)
			if err != nil {
				w.CloseWithError(err)
				return
			}

			encoder := json.NewEncoder(w)
			encoder.SetEscapeHTML(false)
			if err := encoder.Encode(diff); err != nil {
				w.CloseWithError(err)
				return
			}
		}()

		if err := writePrompt(inout.Stdout, r, opts.CommonOptions.Locale); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		if err := r.Close(); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}

		return nil
	}

	diff, err := buildDiff(opts)
	if err != nil {
		return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	switch opts.OutputFormat {
	case OutputFormatDiff:
		if err := diff.Write(inout.Stdout, opts.ShowSame); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	case OutputFormatJSON:
		encoder := json.NewEncoder(inout.Stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(diff); err != nil {
			return fmt.Errorf("cmd.MainCommandByOptions: %w", err)
		}
	}

	if diff.IsEmpty() {
		return nil
	}
	return ErrHasDiff
}

func buildDiff(opts *Options) (*pfd.DiffResult, error) {
	compDelivTable1, err := pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader1)
	if err != nil {
		return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	pA, err := pfdfmt.Parse("1", opts.PFDReader1, &pfdfmt.ParseOptions{
		CompositeDeliverableTable: compDelivTable1,
	}, opts.CommonOptions.Logger)
	if err != nil {
		return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	compDelivTable2, err := pfdtsv.ParseCompositeDeliverableTable(opts.CompositeDeliverableTableReader2)
	if err != nil {
		return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	pB, err := pfdfmt.Parse("2", opts.PFDReader2, &pfdfmt.ParseOptions{
		CompositeDeliverableTable: compDelivTable2,
	}, opts.CommonOptions.Logger)
	if err != nil {
		return nil, fmt.Errorf("cmd.MainCommandByOptions: %w", err)
	}

	return pfd.Diff(pA, pB), nil
}

func writePrompt(w io.Writer, diffReader io.Reader, loc locale.Locale) error {
	switch loc {
	case locale.LocaleJa:
		if _, err := io.WriteString(w, `末尾の DiffResult 型の JSON 文字列で記載された差分を要約し、標準出力へ回答してください：

# 差分データの読み方
## 型
### DiffResult
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .nodeDiff | NodeDiffResult | ノードの差分 |
| .edgeDiff | EdgeDiffResult | 辺の差分 |

### NodeDiffResult
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .extraNodes | []Node | IDが追加されたノード |
| .missingNodes | []Node | IDが削除されたノード |
| .changedNodes | []NodeDiffChange | IDは変わらないが内容が変更されたノード |
| .sameNodes | []Node | IDも内容も変更されていないノード |

### NodeDiffChange
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .id | string | ノードのID |
| .old | []NodeData | 変更前のノードの内容 |
| .new | []NodeData | 変更後のノードの内容 |

### Node
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .id | string | ノードのID |
| .desc | string | ノードの説明 |
| .type | NodeType | ノードの種類 |

### NodeData
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .desc | string | ノードの説明 |
| .type | NodeType | ノードの種類 |

### NodeType
| 値 | 説明 |
| --- | --- |
| "ATOMIC_PROCESS" | 原子プロセス |
| "COMPOSITE_PROCESS" | 複合プロセス |
| "ATOMIC_DELIVERABLE" | 原子成果物 |
| "COMPOSITE_DELIVERABLE" | 複合成果物 |

### EdgeDiffResult
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .extraEdges | []Edge | IDの組が追加された辺 |
| .missingEdges | []Edge | IDの組が削除された辺 |
| .changedEdges | []EdgeDiffChange | IDの組は変わらないがフィードバック辺のフラグが変更された辺 |
| .sameEdges | []Edge | IDの組もフィードバック辺のフラグも変更されていない辺 |

### EdgeDiffChange
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .source | string | 辺のソースID |
| .target | string | 辺のターゲットID |
| .oldFeedbackFlags | []bool | 変更前のフィードバック辺のフラグ |
| .newFeedbackFlags | []bool | 変更後のフィードバック辺のフラグ |

### Edge
| フィールド | 型 | 説明 |
| --- | --- | --- |
| .source | string | 辺のソースID |
| .target | string | 辺のターゲットID |
| .feedback | bool | 辺のフィードバック辺のフラグ |

# 要約のフォーマット
要約ではIDだけだとわかりにくいので、ノードの説明とIDを組み合わせて表示してください。

`+"```text\n"+`
# ノードの変更の概要
{{ .nodeChangeSummary }}

# 辺の変更の概要
{{ .edgeChangeSummary }}
`+"```\n"+`

# 要約すべき差分
`+"```json\n"); err != nil {
			return fmt.Errorf("cmd.writePromptTemplate: %w", err)
		}
		if _, err := io.Copy(w, diffReader); err != nil {
			return fmt.Errorf("cmd.writePromptTemplate: %w", err)
		}
		if _, err := io.WriteString(w, "```\n"); err != nil {
			return fmt.Errorf("cmd.writePromptTemplate: %w", err)
		}
		return nil
	case locale.LocaleEn:
		if _, err := io.WriteString(w, `Summarize the diff in the JSON string at the end of the output as a prompt for Agentic AIs:

# How to read the diff data
## Types
### DiffResult
| Field | Type | Description |
| --- | --- | --- |
| .nodeDiff | NodeDiffResult | Node diff |
| .edgeDiff | EdgeDiffResult | Edge diff |

### NodeDiffResult
| Field | Type | Description |
| --- | --- | --- |
| .extraNodes | []Node | Nodes added |
| .missingNodes | []Node | Nodes removed |
| .changedNodes | []NodeDiffChange | Nodes with changed content |
| .sameNodes | []Node | Nodes with no changes |

### NodeDiffChange
| Field | Type | Description |
| --- | --- | --- |
| .id | string | Node ID |
| .old | []NodeData | Old node data |
| .new | []NodeData | New node data |

### Node
| Field | Type | Description |
| --- | --- | --- |
| .id | string | Node ID |
| .desc | string | Node description |
| .type | NodeType | Node type |

### NodeData
| Field | Type | Description |
| --- | --- | --- |
| .desc | string | Node description |
| .type | NodeType | Node type |

### NodeType
| Value | Description |
| --- | --- |
| "ATOMIC_PROCESS" | Atomic process |
| "COMPOSITE_PROCESS" | Composite process |
| "ATOMIC_DELIVERABLE" | Atomic deliverable |
| "COMPOSITE_DELIVERABLE" | Composite deliverable |

### EdgeDiffResult
| Field | Type | Description |
| --- | --- | --- |
| .extraEdges | []Edge | Edges added |
| .missingEdges | []Edge | Edges removed |
| .changedEdges | []EdgeDiffChange | Edges with changed content |
| .sameEdges | []Edge | Edges with no changes |

### EdgeDiffChange
| Field | Type | Description |
| --- | --- | --- |
| .source | string | Edge source ID |
| .target | string | Edge target ID |
| .oldFeedbackFlags | []bool | Old feedback flags |
| .newFeedbackFlags | []bool | New feedback flags |

### Edge
| Field | Type | Description |
| --- | --- | --- |
| .source | string | Edge source ID |
| .target | string | Edge target ID |
| .feedback | bool | Edge feedback flag |

# Summary format
To summarize the diff, please combine the node description and ID.

`+"```text\n"+`
# Node change summary
{{ .nodeChangeSummary }}

# Edge change summary
{{ .edgeChangeSummary }}
`+"```\n"+`

# Summary should be included diff
`+"```json\n"); err != nil {
			return fmt.Errorf("cmd.writePromptTemplate: %w", err)
		}
		if _, err := io.Copy(w, diffReader); err != nil {
			return fmt.Errorf("cmd.writePromptTemplate: %w", err)
		}
		if _, err := io.WriteString(w, "```\n"); err != nil {
			return fmt.Errorf("cmd.writePromptTemplate: %w", err)
		}
		return nil
	default:
		panic(fmt.Sprintf("cmd.writePromptTemplate: unknown locale: %s", loc))
	}
}
