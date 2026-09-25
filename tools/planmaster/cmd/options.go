package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/Kuniwak/pfd-tools/cli"
	"github.com/Kuniwak/pfd-tools/masterschedule"
	"github.com/Kuniwak/pfd-tools/mastertsv"
	"github.com/Kuniwak/pfd-tools/slograw"
	"github.com/Kuniwak/pfd-tools/tools"
)

type Options struct {
	CommonOptions           *tools.CommonOptions
	BusinessTimeFuncOptions *tools.BusinessTimeFuncOptions
	BufferMultiplier        float64
	WriteMasterSchedule     masterschedule.Writer
	Logger                  *slog.Logger
	PlanReader              io.Reader
	MasterTSVReader         io.Reader

	RowDescriptions map[string]string
	BarDescriptions map[string]string
}

func ParseOptions(args []string, inout *cli.ProcInout) (*Options, error) {
	flags := flag.NewFlagSet("planmaster", flag.ContinueOnError)
	flags.SetOutput(inout.Stderr)
	flags.Usage = func() {
		tools.PrintUsageHeader(flags, "Usage: planmaster [options] -tsv <master-tsv> [-b <buffer-multiplier>] <plan>", ShortHelp)
		fmt.Fprintf(flags.Output(), `
Specification
    マスタースケジュールは、実行計画の原子プロセスを分類表（-tsv）で行（MasterRow）と
    バー（MasterBar）に束ねた図表です。

    分類表（TSV）:
      - ID・MasterRow・MasterBar 列を名前で同定します（列順は自由。他の列は無視します）。
      - MasterRow はカンマ区切りで複数指定できます。MasterBar は 1 つだけです。
      - 同じ ID を複数行に書くとエラーです。実行計画にない ID もエラーです。
      - すべての原子プロセスを分類する必要はありません。分類しなかった原子プロセスは
        描画されません（表を絞って一部だけ描けます。網羅したいときの検査は別ツールの
        責務です）。ID・MasterRow・MasterBar のいずれかが空欄の行も「分類なし」として描画しません。
      - どの軸で分類するかは分類表を作る側が決めます。原子プロセス表（ap.tsv）に
        MasterRow / MasterBar 列を書けば、未知の列は無視されるため ap.tsv をそのまま分類表として
        渡せます。別の軸で描くときは別の分類表を作って渡します。

    描画規則:
      - 各原子プロセスは初回実行の区間だけを描画対象にします。手戻り（2 周目以降の実行）
        は描画しません。締切として意味を持つのは初回完了であり、手戻り込みの遅い終了を
        見せるとそれに合わせて動かれてしまうためです（パーキンソンの法則）。
      - (MasterRow, MasterBar) のバーは、そこに分類された原子プロセスの初回実行区間の
        [最小の開始, 最大の終了] です。

    メタ表（-row-meta / -bar-meta。省略可）:
      - ID 列と説明列（Description または 説明）を名前で同定し、他の列は無視します。
        既存のグループ表・マイルストーン表をそのまま渡せます。
      - 省略した場合、説明は空欄になります。

Example
    $ planmaster -f path/to/project.json -tsv path/to/master.tsv path/to/plan.json
	MasterRow       MasterRowDescription    MasterBar       MasterBarDescription    Start   End
	G1              M1              2025-11-18 10:00:00     2025-11-21 14:30:00
	...

    # MasterRow / MasterBar 列を書いた ap.tsv をそのまま分類表として渡す
    $ planmaster -f path/to/project.json -tsv path/to/atomic_proc.tsv -row-meta path/to/group.tsv -bar-meta path/to/milestone.tsv path/to/plan.json
	MasterRow       MasterRowDescription    MasterBar       MasterBarDescription    Start   End
	G1      グループ1       M1      マイルストーン1 2025-11-18 10:00:00     2025-11-21 14:30:00
	...

    # クリティカルパス上の原子プロセスを含むバー（マイルストーン）を強調する
    # （強調 ID 表はバーの ID の表なので、criticalpath の出力を分類表と qhs で突き合わせて作る）
    $ criticalpath -poor -f path/to/project.json >cp.tsv
    $ qhs -H -O -t -T 'SELECT DISTINCT t1.MasterBar AS ID FROM path/to/atomic_proc.tsv AS t1 JOIN cp.tsv AS t2 ON t1.ID = t2.ID WHERE t2."最大弾性値（全余裕）" < 0.0001 AND LENGTH(t1.MasterBar) > 0' >em.tsv
    $ planmaster -f path/to/project.json -tsv path/to/atomic_proc.tsv -out-format mermaid -em-tsv em.tsv path/to/plan.json
	gantt
	    dateFormat YYYY-MM-DD HH:mm
	    section G1 グループ
	    M1 マイルストーン1 :crit, 2025-11-18 10:00, 2025-11-21 14:30
	    M2 マイルストーン2 :2025-11-21 14:30, 2025-11-25 14:30
	...
`)
	}

	var commonRawOptions tools.CommonRawOptions
	tools.DeclareCommonOptions(flags, &commonRawOptions)

	var configShortPath, configLongPath string
	tools.DeclareConfigOptions(flags, &configShortPath, &configLongPath)

	var masterTSVPath string
	flags.StringVar(&masterTSVPath, "tsv", "", "path to the master TSV (columns: ID, MasterRow, MasterBar)")
	var rowMetaPath, barMetaPath string
	flags.StringVar(&rowMetaPath, "row-meta", "", "path to the row description TSV (columns: ID, Description or 説明)")
	flags.StringVar(&barMetaPath, "bar-meta", "", "path to the bar description TSV (columns: ID, Description or 説明)")

	var planOutputFormatRawOptions tools.PlanOutputFormatRawOptions
	tools.DeclareBusinessTimeFuncOptions(flags, &planOutputFormatRawOptions.BusinessTimeFuncRawOptions)

	flags.StringVar(&planOutputFormatRawOptions.OutputFormat, tools.OutFormatFlag, "", "output format (available: google-spreadsheet-tsv, mermaid, plantuml. default: google-spreadsheet-tsv)")

	tools.DeclareEmphasisTSVOptions(flags, &planOutputFormatRawOptions.EmphasisTSVPath)

	var bufferMultiplierLong, bufferMultiplierShort float64
	flags.Float64Var(&bufferMultiplierShort, "b", -1.0, "plan multiplier (negative value means not specified. 1.0 if both -b and -buffer-multiplier are not specified)")
	flags.Float64Var(&bufferMultiplierLong, "buffer-multiplier", -1.0, "plan multiplier (negative value means not specified. 1.0 if both -b and -buffer-multiplier are not specified)")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Options{CommonOptions: &tools.CommonOptions{Help: true}}, nil
		}
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	commonOptions, err := tools.ValidateCommonOptions(&commonRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if commonOptions.ShortHelp {
		return &Options{CommonOptions: commonOptions}, nil
	}
	if commonOptions.Version {
		return &Options{CommonOptions: commonOptions}, nil
	}

	var bufferMultiplier float64
	if bufferMultiplierLong >= 0 {
		bufferMultiplier = bufferMultiplierLong
	} else if bufferMultiplierShort >= 0 {
		bufferMultiplier = bufferMultiplierShort
	} else {
		bufferMultiplier = 1.0
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	var fsmRawOptions tools.FSMRawOptions
	projectConfig, _, err := tools.ReadProjectConfig(&configShortPath, &configLongPath, fsmRawOptions, planOutputFormatRawOptions, cwd, flags)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	em, err := tools.ValidateEmphasisTSVOptions(projectConfig.EmphasisTSVPath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	logger := slog.New(slograw.NewHandler(inout.Stderr, commonOptions.LogLevel))

	writeMasterSchedule, err := masterschedule.ParseOutputFormat(projectConfig.OutputFormat, em, logger)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	businessTimeFuncOptions, err := tools.ValidateBusinessTimeFuncOptions(&projectConfig.BusinessTimeFuncRawOptions)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if masterTSVPath == "" {
		return nil, fmt.Errorf("cmd.ParseOptions: -tsv is required")
	}
	masterTSVReader, err := os.Open(tools.ResolvePath(cwd, masterTSVPath))
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	rowDescriptions, err := ParseDescriptionsPath(cwd, rowMetaPath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}
	barDescriptions, err := ParseDescriptionsPath(cwd, barMetaPath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	if flags.NArg() > 1 {
		return nil, fmt.Errorf("cmd.ParseOptions: too many arguments: expected 1 plan, got %d", flags.NArg())
	}
	planPath := projectConfig.PlanPath
	if flags.Arg(0) != "" {
		planPath = tools.ResolvePath(cwd, flags.Arg(0))
	}
	if planPath == "" {
		return nil, fmt.Errorf("cmd.ParseOptions: plan path is required")
	}

	planReader, err := os.Open(planPath)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseOptions: %w", err)
	}

	return &Options{
		CommonOptions:           commonOptions,
		BusinessTimeFuncOptions: businessTimeFuncOptions,
		BufferMultiplier:        bufferMultiplier,
		WriteMasterSchedule:     writeMasterSchedule,
		Logger:                  logger,
		PlanReader:              planReader,
		MasterTSVReader:         masterTSVReader,
		RowDescriptions:         rowDescriptions,
		BarDescriptions:         barDescriptions,
	}, nil
}

func ParseDescriptionsPath(cwd string, path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}
	f, err := os.Open(tools.ResolvePath(cwd, path))
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseDescriptionsPath: %w", err)
	}
	defer f.Close()
	descs, err := mastertsv.ParseDescriptions(f)
	if err != nil {
		return nil, fmt.Errorf("cmd.ParseDescriptionsPath: %w", err)
	}
	return descs, nil
}
