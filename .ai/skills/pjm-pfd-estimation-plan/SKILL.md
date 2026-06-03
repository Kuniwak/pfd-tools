---
name: pjm-pfd-estimation-plan
description: |
  A workflow that uses pfd-tools to produce a two-point optimistic/pessimistic estimate from the ap.tsv of a bootstrapped project. It derives scenario tables, lints, runs pfdplan, analyzes the critical path, runs an improvement loop, and generates a Mermaid Gantt or Google Sheets Timeline. Triggers when the user says things like "run the estimate", "I want to build a Gantt", "check the critical path", "generate a timeline" (or in Japanese, "見積もりを実行", "Gantt を作りたい", "クリティカルパスを確認", "タイムラインを生成").
allowed-tools: Read, Write, Bash, Glob, Grep, readfile, writefile, terminal, glob, grep
---

# pjm-pfd-estimation-plan

Run the two-point estimate for a project created with `pjm-pfd-project-bootstrap`. The only file you edit is the master `<project>/ap.tsv`.

## Workflow

- [ ] 1. Confirm the working language: ask the user whether to proceed in 日本語 or English. Use the chosen language for all subsequent conversation, explanations, progress reports, and lint / critical-path summaries. (Command names, paths, option flags, column names, and code stay verbatim regardless of the chosen language.)
- [ ] 2. Read `<project>/answers.md` ([Bootstrap Answers](#bootstrap-answers))
- [ ] 3. Enter estimates into the master `ap.tsv` ([Master AP Table](#master-ap-table))
- [ ] 4. (`resource_conflicts: disabled` only) Deconflict resources ([Optional Resource Deconfliction](#optional-resource-deconfliction))
- [ ] 5. Derive the optimistic/pessimistic scenario tables ([Derive Scenario Tables](#derive-scenario-tables))
- [ ] 6. Update the resource tables (when resources change) → lint both scenarios ([Project Files And Lint](#project-files-and-lint))
- [ ] 7. Confirm the start date ([Plan Start](#plan-start))
- [ ] 8. Generate output ([GitHub + Mermaid](#github--mermaid) or [Google Sheets Timeline](#google-sheets-timeline))
- [ ] 9. Analyze the critical path and propose improvements ([Critical Path Review](#critical-path-review) → [Improvement Loop](#improvement-loop))

## Conventions

- At the start, confirm the working language (日本語 / English) with the user, then use it consistently for all conversation, explanations, progress reports, and lint summaries. Keep command names, paths, options, column names, and code verbatim.
- The master `<project>/ap.tsv` is the human-edited source data and is not passed to pfd-tools. The inputs to pfd-tools (pfdlint / pfdtable / pfdplan / criticalpath) are always the `ap_optimistic.tsv` / `ap_pessimistic.tsv` derived with qhs (single `Needed Resources` column) and `project_optimistic.json` / `project_pessimistic.json`.
- The resource tables are per-scenario: `r_optimistic.tsv` / `r_pessimistic.tsv` (the resource ID set differs when the optimistic and pessimistic worker counts differ). Generate each table from its corresponding derived table: `pfdtable -t r -locale en -p <project>/pfd.drawio.png -ap <project>/ap_optimistic.tsv -cd <project>/cd.tsv > <project>/r_optimistic.tsv` (for pessimistic, `ap_pessimistic.tsv` → `r_pessimistic.tsv`). Since the output name differs from the input, `>` is fine.
- The prerequisite is a bootstrapped project. If running directly, verify pfd-tools 5.1.0 or later and qhs (if missing, see Tool Setup in the bootstrap skill).

## Bootstrap Answers

Before changing any table, read `<project>/answers.md`. Bootstrap writes the following:

```yaml
project: <project>
output_format: github-mermaid        # or google-sheets-timeline
resource_conflicts: enabled          # if disabled, perform Deconfliction
optimistic_worker_count: 2           # filled in when enabled (fine granularity); omitted when disabled
pessimistic_worker_count: 2          # same as above
plan_start_date: 2026-05-20
```

If `answers.md` is missing, or `resource_conflicts` is missing, default to `enabled` and tell the user. Confirm and update `plan_start_date` in [Plan Start](#plan-start).

## Master AP Table

The only file the user edits is `<project>/ap.tsv`. **Do not change values the user has already entered** (only blanks get a default at derivation time). Header:

```tsv
ID	Description	Optimistic Work Volume	Pessimistic Work Volume	Optimistic Rework Volume Ratio	Pessimistic Rework Volume Ratio	Optimistic Needed Resources	Pessimistic Needed Resources	Start Condition	Group	Estimation Basis
```

| Column | Meaning | Default / Notes |
|---|---|---|
| `Optimistic Work Volume` | Work volume when nothing goes wrong (unit: business days) | — |
| `Pessimistic Work Volume` | Work volume the owner can responsibly commit to (unit: business days) | — |
| `Optimistic Rework Volume Ratio` | Optimistic rework recovery ratio X (0.0–1.0). With work volume V and n rework cycles, the recovered amount is `V * X^n` | Blank → `0.1` at derivation |
| `Pessimistic Rework Volume Ratio` | Pessimistic recovery ratio X (same formula) | Blank → `0.3` at derivation |
| `Optimistic Needed Resources` | Optimistic needed resources | Blank → filled with a default at derivation. Fine granularity: `optimistic_worker_count` workers; coarse granularity: `DeptA:1` etc. |
| `Pessimistic Needed Resources` | Pessimistic needed resources | Blank → filled with a default at derivation. Fine granularity: `pessimistic_worker_count` workers; coarse granularity: same value in both columns |
| `Start Condition` | Preceding deliverable | Initially blank |
| `Group` | Group of the Google Sheets Timeline card | — |
| `Estimation Basis` | A short rationale for the estimate | — |

If your table's Description column uses a different header name, adjust the qhs query to match the actual header.

## Optional Resource Deconfliction

Only when `resource_conflicts: disabled`, before deriving scenarios, rewrite `Optimistic Needed Resources` / `Pessimistic Needed Resources` in the master `ap.tsv` so each resource occurrence is treated independently. Re-read `answers.md` right before this (it can change midway).

Rules:

- Rewrite only the two columns `Optimistic Needed Resources` / `Pessimistic Needed Resources`. Preserve other cells in a TSV-aware way (a whole-file regex replace is not allowed).
- Preserve `;` (OR) and `,` (AND resource group). Split each assignment only at the first `:`, and leave the consumed-work-volume text as-is (do not split, round, or normalize it).
- Mechanically append `_<process_id>` to the resource ID (if it already ends with `_<process_id>`, do nothing).
- Normalize rewritten cells to the compact syntax (`Foo: 1; Bar:0.5` → `Foo1:1;Bar1:0.5`).

```text
Foo:1        -> Foo_123:1
Foo,Bar:1    -> Foo_123,Bar_123:1
Foo:1;Bar:1  -> Foo_123:1;Bar_123:1
Foo_123:1    -> Foo_123:1
```

Tell the user the rewritten resource IDs. Then proceed to [Derive Scenario Tables](#derive-scenario-tables) → [Project Files And Lint](#project-files-and-lint), regenerating the resource tables (`r_optimistic.tsv` / `r_pessimistic.tsv`) after derivation. When `resource_conflicts: enabled`, do not change the two columns.

## Derive Scenario Tables

`AS` `Optimistic Work Volume` / `Pessimistic Work Volume` to `Est. Work Volume`, and `Optimistic Needed Resources` / `Pessimistic Needed Resources` to `Needed Resources`. To fill blank rework ratios with defaults (optimistic `0.1` / pessimistic `0.3`), use `CASE WHEN LENGTH(...) > 0 THEN ... ELSE <default> END` (this avoids writing the empty-string literal `''` and avoids shell quoting).

Blank `Optimistic Needed Resources` / `Pessimistic Needed Resources` get a default at derivation so the `Needed Resources` column is always non-empty (passing it empty to pfd-tools makes pfdlint error out; filling every cell by hand is a heavy burden, so fill with a plausible default and prompt the user to fix it afterward).

Default value: for fine granularity (`enabled`), `optimistic_worker_count` / `pessimistic_worker_count` workers as `WorkerA:1;WorkerB:1;...;WorkerN:1` (`1` → `WorkerA:1`, `2` → `WorkerA:1;WorkerB:1`; the one after `WorkerZ` is `WorkerAA`). For coarse granularity (`disabled`), `DeptA:1`. Worker counts are assumed to be at least 1 for both optimistic and pessimistic (a count of 0 would make the default empty). Replace `WorkerA:1;WorkerB:1;...;WorkerN:1` / `WorkerA:1;WorkerB:1;...;WorkerM:1` in the queries below with this default.

```console
$ qhs -H -O -t -T "SELECT ID, Description, \"Optimistic Work Volume\" AS \"Est. Work Volume\", CASE WHEN LENGTH(\"Optimistic Rework Volume Ratio\") > 0 THEN \"Optimistic Rework Volume Ratio\" ELSE 0.1 END AS \"Est. Rework Volume Ratio\", CASE WHEN LENGTH(\"Optimistic Needed Resources\") > 0 THEN \"Optimistic Needed Resources\" ELSE 'WorkerA:1;WorkerB:1;...;WorkerN:1' END AS \"Needed Resources\", \"Start Condition\", \"Group\", \"Estimation Basis\" FROM \"-\"" < <project>/ap.tsv > <project>/ap_optimistic.tsv
$ qhs -H -O -t -T "SELECT ID, Description, \"Pessimistic Work Volume\" AS \"Est. Work Volume\", CASE WHEN LENGTH(\"Pessimistic Rework Volume Ratio\") > 0 THEN \"Pessimistic Rework Volume Ratio\" ELSE 0.3 END AS \"Est. Rework Volume Ratio\", CASE WHEN LENGTH(\"Pessimistic Needed Resources\") > 0 THEN \"Pessimistic Needed Resources\" ELSE 'WorkerA:1;WorkerB:1;...;WorkerM:1' END AS \"Needed Resources\", \"Start Condition\", \"Group\", \"Estimation Basis\" FROM \"-\"" < <project>/ap.tsv > <project>/ap_pessimistic.tsv
```

Tell the user about the cells filled with defaults and have them fix as needed.

In PowerShell, replace each line with `Get-Content <project>/ap.tsv | qhs -H -O -t -T '<same SQL>' > <out>` (the alternative to `< <in>`).

## Project Files And Lint

`project_optimistic.json` / `project_pessimistic.json` are already copied by bootstrap from the scaffold (`assets/`), with `atomic_process_table` and `resource_table` pointing to `ap_optimistic.tsv` + `r_optimistic.tsv` / `ap_pessimistic.tsv` + `r_pessimistic.tsv` respectively (`atomic_deliverable_table` and `composite_deliverable_table` are shared by both). Since derivation overwrites the same-named tables, no re-creation is needed.

If you performed Deconfliction or changed resources, regenerate the resource tables (see [Conventions](#conventions)) before linting:

```console
$ pfdtable -t r -locale en -p <project>/pfd.drawio.png -ap <project>/ap_optimistic.tsv -cd <project>/cd.tsv > <project>/r_optimistic.tsv
$ pfdtable -t r -locale en -p <project>/pfd.drawio.png -ap <project>/ap_pessimistic.tsv -cd <project>/cd.tsv > <project>/r_pessimistic.tsv
$ pfdlint -f <project>/project_optimistic.json
$ pfdlint -f <project>/project_pessimistic.json
```

Report the lint results starting from the error / warning counts, listing each issue's severity, target ID, an easy-to-understand explanation of the message, and an improvement suggestion. Do not run `pfdplan` while errors remain. Keep warnings only if the PjM can explain them.

## Plan Start

Read `plan_start_date` from `<project>/answers.md`.

- If present: present the existing date and confirm whether to use it as-is or a different date (YYYY-MM-DD). On change, update `answers.md`.
- If absent (old answers.md): confirm a new entry in YYYY-MM-DD and append it to `answers.md`.

This date becomes the first task's start date on the Mermaid Gantt / Google Sheets Timeline.

## GitHub + Mermaid

When `output_format: github-mermaid`:

```console
$ pfdplan -f <project>/project_optimistic.json -poor -out-format mermaid -start <yyyy-mm-dd> -start-time 10:00 -duration 8 -not-biz-days <(holidays -locale ja) > <project>/plan_optimistic.mmd
$ pfdplan -f <project>/project_pessimistic.json -poor -out-format mermaid -start <yyyy-mm-dd> -start-time 10:00 -duration 8 -not-biz-days <(holidays -locale ja) > <project>/plan_pessimistic.mmd
```

`<(holidays -locale ja)` is Bash process substitution. It is unavailable on Windows / PowerShell / `sh`, so first generate `holidays -locale ja > <project>/holidays_ja.txt` and pass `-not-biz-days <project>/holidays_ja.txt` (the same applies to the `pfdplan` under Google Sheets Timeline below).

Keep the `.mmd` as the preview target. Do not create a separate `plan.md`.

## Google Sheets Timeline

When `output_format: google-sheets-timeline`, generate the timeline and critical-path TSVs:

```console
$ pfdplan -f <project>/project_optimistic.json -poor -out-format google-spreadsheet-tsv -start <yyyy-mm-dd> -start-time 10:00 -duration 8 -not-biz-days <(holidays -locale ja) > <project>/timeline_optimistic.tsv
$ pfdplan -f <project>/project_pessimistic.json -poor -out-format google-spreadsheet-tsv -start <yyyy-mm-dd> -start-time 10:00 -duration 8 -not-biz-days <(holidays -locale ja) > <project>/timeline_pessimistic.tsv
$ criticalpath -f <project>/project_optimistic.json -poor > <project>/criticalpath_optimistic.tsv
$ criticalpath -f <project>/project_pessimistic.json -poor > <project>/criticalpath_pessimistic.tsv
```

Sheets to create:

| Sheet | Content |
|---|---|
| `Optimistic Timeline` / `Pessimistic Timeline` | Timeline view sourced from `Optimistic Data` / `Pessimistic Data` |
| `Optimistic Data` / `Pessimistic Data` | Paste `timeline_*.tsv` and add the `Group` and `Critical Path` columns |
| `Optimistic Critical Path Analysis` / `Pessimistic Critical Path Analysis` | Paste `criticalpath_*.tsv` and add column D `Critical Path` |
| `Atomic Process Table` | Paste `ap.tsv` (the lookup source for `Group`) |

Change the A–H header of `Optimistic Data` / `Pessimistic Data`:

```tsv
Atomic Process	Completions	Allocated Resources	Description	Start Time	End Time	Start	End
```

<details>
<summary>Formulas / conditional formatting / Timeline view settings to add</summary>

I1 of `Optimistic Data` / `Pessimistic Data` (`Group`):

```text
={"Group";ARRAYFORMULA(IF(ISBLANK(A2:A),"",VLOOKUP(A2:A,'Atomic Process Table'!A2:J,9,FALSE)))}
```

D1 of each critical-path analysis sheet:

```text
={"Critical Path";ARRAYFORMULA(IF(ISBLANK(A2:A),"",IF(B2:B<0.0001,TRUE,FALSE)))}
```

J1 of `Optimistic Data` / J1 of `Pessimistic Data` (point each reference at its own analysis sheet):

```text
={"Critical Path";ARRAYFORMULA(IF(ISBLANK(A2:A),"",VLOOKUP(A2:A,'Optimistic Critical Path Analysis'!A2:D,4,FALSE)))}
={"Critical Path";ARRAYFORMULA(IF(ISBLANK(A2:A),"",VLOOKUP(A2:A,'Pessimistic Critical Path Analysis'!A2:D,4,FALSE)))}
```

Timeline view settings: data range `Optimistic Data!A1:J` / `Pessimistic Data!A1:J`, start date `Start Time`, end date or duration `End Time`, card title `Atomic Process`, color `Critical Path`, details `Description`, group `Group`.

Conditional formatting: range `A2:J`, custom formula `=$J2=TRUE`, fill light pink.

</details>

## Critical Path Review

Start improvement consideration from the pessimistic case. If `criticalpath_pessimistic.tsv` does not exist, generate it (reuse it if already generated in Google Sheets Timeline mode):

```console
$ criticalpath -f <project>/project_pessimistic.json -poor > <project>/criticalpath_pessimistic.tsv
```

Explain the processes with `TOTAL_FLOAT < 0.0001` and the processes whose `MINIMUM_ELASTICITY` is not `-`.

## Improvement Loop

You can apply one of the following improvements to the critical path. Propose the ones that look applicable from the current critical-path analysis:

1. Parallelize serial processes on the critical path
  * Note that successor processes often become speculative because they can no longer use the intermediate deliverable
2. Make the critical path's input deliverable an initial deliverable
  * Buy an off-the-shelf one (cost goes up)
3. Shorten the lead time of an atomic process on the critical path
  * Efficiency gains such as automation
  * Increase consumed work volume by dividing the work among people
  * Buy non-dedicated equipment (cost goes up)
4. Add resources
  * Add more personnel (cost goes up)
  * Buy dedicated equipment (cost goes up)

Then re-derive from [Derive Scenario Tables](#derive-scenario-tables); when resources change, regenerate the resource tables (`r_optimistic.tsv` / `r_pessimistic.tsv`) ([Conventions](#conventions)), and finally regenerate the lint, plan, and critical-path output.
