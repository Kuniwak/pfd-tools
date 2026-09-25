---
name: pjm-pfd-estimation-plan
description: |
  A workflow that uses pfd-tools to produce a two-point optimistic/pessimistic estimate from the estimate boxes on the PFD and the ap.tsv of a bootstrapped project. It derives scenario tables, lints, runs pfdplan, analyzes the critical path, runs an improvement loop, and generates a Mermaid Gantt or Google Sheets Timeline. Triggers when the user says things like "run the estimate", "I want to build a Gantt", "check the critical path", "generate a timeline" (or in Japanese, "見積もりを実行", "Gantt を作りたい", "クリティカルパスを確認", "タイムラインを生成").
allowed-tools: Read, Write, Bash, Glob, Grep, readfile, writefile, terminal, glob, grep
---

# pjm-pfd-estimation-plan

Run the two-point estimate for a project created with `pjm-pfd-project-bootstrap`. The user enters the estimates (optimistic/pessimistic work volume) in the estimate boxes on the PFD; everything else (needed resources, etc.) is edited in the master `<project>/ap.tsv`. If along the way you want to check PFD diffs, successor processes, and so on, pfd-tools can help. Run `pfdhelp -short` to see a description of the pfd-tools tools.

## Workflow

- [ ] 1. Confirm the working language: ask the user whether to proceed in 日本語 or English. Use the chosen language for all subsequent conversation, explanations, progress reports, and lint / critical-path summaries. (Command names, paths, option flags, column names, and code stay verbatim regardless of the chosen language.)
- [ ] 2. Read the scenario settings files ([Scenario Settings](#scenario-settings))
- [ ] 3. Number and fix the PFD and check its soundness (`pfdrenum` / `pfdfix` / `pfdestcallout` / `pfdlint`) ([PFD Fix](#pfd-fix))
- [ ] 4. Have the user fill in the estimate boxes on the PFD, and prepare the master `ap.tsv` ([Estimate Boxes](#estimate-boxes) / [Master AP Table](#master-ap-table))
- [ ] 5. Derive the atomic process table for each scenario ([Derive Scenario Tables](#derive-scenario-tables))
- [ ] 6. Update the resource tables (finite resource model, when resources change) → lint each scenario ([Project Files And Lint](#project-files-and-lint))
- [ ] 7. Confirm the start date ([Plan Start](#plan-start))
- [ ] 8. Generate the execution plan plan-json ([Generate Plan JSON](#generate-plan-json))
- [ ] 9. Generate output ([GitHub + Mermaid](#github--mermaid) or [Google Sheets Timeline](#google-sheets-timeline)). Optionally a master schedule ([Master Schedule](#master-schedule))
- [ ] 10. Analyze the critical path and propose improvements ([Critical Path Review](#critical-path-review) → [Improvement Loop](#improvement-loop))

## Conventions

- At the start, confirm the working language (日本語 / English) with the user, then use it consistently for all conversation, explanations, progress reports, and lint summaries. Keep command names, paths, options, column names, and code verbatim.
- The source data for the estimates (optimistic/pessimistic work volume) is the estimate boxes on the PFD, which `pfdesttable` reads out as a table. The master `<project>/ap.tsv` is human-edited source data (columns that cannot live on the PFD, such as needed resources) and is not passed to pfd-tools. The inputs to pfd-tools (pfdlint / pfdtable / pfdplan / criticalpath) are always `ap_optimistic.tsv` / `ap_pessimistic.tsv` (with a single `必要資源` (needed resources) column), derived by JOINing the `pfdesttable` output with the master using qhs, and `project_optimistic.json` / `project_pessimistic.json`.
- There is one scenario settings file per scenario (`project_<scenario>.json`). This is not limited to the two optimistic/pessimistic points: multi-point estimates such as a three-point estimate (optimistic, most likely, pessimistic) are handled just by adding points to the estimate boxes and adding scenario settings files. The steps below use a two-point estimate as the example, but with more scenarios you repeat the same steps.
- The execution model is determined by `resource_mode` (`finite` / `infinite`) and `feedback_mode` (`enabled` / `disabled`) in each scenario settings file. With `resource_mode: infinite`, resource conflicts are not considered, so the `必要資源` column and the resource tables are not used (the result is the same as every process having its own resource). With `feedback_mode: disabled`, rework is not considered, so the `予想手戻り作業量割合` (expected rework volume ratio) column and the `最大版` (max revision) column are not used.
- Only in the finite resource model (`resource_mode: finite`), keep the resource tables per scenario as `r_optimistic.tsv` / `r_pessimistic.tsv` (the resource ID set differs when the worker count differs between scenarios). Generate each table from its corresponding derived table: `pfdtable -t r -res finite -p <project>/pfd.drawio.png -ap <project>/ap_optimistic.tsv -cd <project>/cd.tsv > <project>/r_optimistic.tsv` (for pessimistic, `ap_pessimistic.tsv` → `r_pessimistic.tsv`). Since the output name differs from the input, `>` is fine.
- The prerequisite is a bootstrapped project. If running directly, verify that pfd-tools supports `pfdesttable` and the execution-model selection (`-res` / `-fb`), and that qhs is installed (if the output of `pfdplan -h` contains `-res`, it is supported; if anything is missing, see Tool Setup in the bootstrap skill).
- <a id="quality-fallback"></a>Search-quality fallback: `pfdplan` (plan-json generation) and `criticalpath`, which perform a search, use `-poor` by default. If `-poor` does not finish within 20 seconds, abort it and rerun the same command with `-poorest` (`-poorest` is greedy and deterministic, and finishes in a practical time even at scales where `-poor` does not complete). So that optimistic and pessimistic can be compared at the same search quality, once you switch to `-poorest` for either one, run the other and all subsequent `pfdplan` / `criticalpath` with `-poorest` as well. `plantimeline` / `planmaster` only render a plan-json and do not search, so they are not affected.

## Scenario Settings

Before changing any table, read all `<project>/project_*.json`. Bootstrap writes the pfd-tools settings and the estimation assumptions together into one file:

```json
{
  "pfd": "pfd.drawio.png",
  "atomic_process_table": "ap_optimistic.tsv",
  "atomic_deliverable_table": "ad.tsv",
  "composite_deliverable_table": "cd.tsv",
  "resource_table": "r_optimistic.tsv",
  "milestone_table": "m.tsv",
  "group_table": "g.tsv",
  "resource_mode": "finite",
  "feedback_mode": "disabled",
  "scenario": "optimistic",
  "output_format": "mermaid",
  "start_day": "2026-05-20",
  "duration": 8,
  "not_biz_days": "holidays_ja.txt",
  "worker_count": 2
}
```

Read `resource_mode` (whether to consider resource conflicts), `feedback_mode` (whether to consider rework), `output_format` (output format), `start_day` (start date), and `worker_count` (the worker count for that scenario; `finite` only).

`output_format`, `start_day`, `duration`, `not_biz_days`, `resource_mode`, and `feedback_mode` are keys that pfd-tools itself reads, so there is no need to pass them again on the `pfdplan` / `plantimeline` command line (if you do, the command line takes precedence). `scenario` and `worker_count` are keys used only by the skill; pfd-tools ignores unknown keys, so they can coexist.

If `resource_mode` / `feedback_mode` is missing, the pfd-tools defaults (`infinite` / `disabled`) apply. Tell the user so. Confirm and update `start_day` in [Plan Start](#plan-start).

### Old-format projects

If `<project>/answers.md` exists, the estimation assumptions are still in the old format (pfd-tools 6.6.x or earlier). **This skill does not read `answers.md`.** Tell the user to migrate first with `pjm-pfd-migrate-v7` and then come back.

## PFD Fix

Users often edit the PFD after bootstrap. Before derivation, lint, and pfdplan, number and fix the PFD and re-check its soundness.

1. **If the PFD was edited, run `pfdrenum`** to number unnumbered nodes (check without `-inplace` → then `-inplace`):

   ```console
   $ pfdrenum <project>/pfd.drawio.png
   $ pfdrenum -inplace <project>/pfd.drawio.png
   ```

2. **`pfdfix`** connects "edges that look like they touch a rectangle but have no `source`/`target` set" (check without `-inplace` → then `-inplace`):

   ```console
   $ pfdfix <project>/pfd.drawio.png
   $ pfdfix -inplace <project>/pfd.drawio.png
   ```

   By default, `pfdfix` does not delete edges whose endpoint falls in no rectangle (unhit) or that would become self-loops; it leaves them with a WARN. **Do not pass `-delete-unhit` unless the user explicitly instructs it.** If a WARN appears, tell the user that the connection target is ambiguous or unconnected, and have them decide whether to revise the drawing on the PFD side.

3. **`pfdestcallout`** fills in the scaffolding for the estimate boxes (idempotent: placeholders are added only directly below added processes, and existing entered values are preserved; check without `-inplace` → then `-inplace`):

   ```console
   $ pfdestcallout <project>/pfd.drawio.png
   $ pfdestcallout -inplace <project>/pfd.drawio.png
   ```

4. **Re-check soundness.** `pfdrenum` / `pfdfix` (and PFD edits) change the graph's connections, so use `pfdlint` to check that **no new soundness violations (cycles, single-src violations, etc.) have appeared**. If you changed the PFD's structure (added or removed processes / deliverables), first update the masters `ap.tsv` / `ad.tsv` / `cd.tsv` and re-derive in [Derive Scenario Tables](#derive-scenario-tables) before linting. If the derived tables are up to date (structure unchanged), you can check right away:

   ```console
   $ pfdlint -f <project>/project_optimistic.json
   ```

   If violations appear, resolve them before moving on (they are checked again by the lint in [Project Files And Lint](#project-files-and-lint)).

## Estimate Boxes

**The user fills in** the estimates in the estimate boxes on the PFD (the placeholder `楽観: d 悲観: d` that `pfdestcallout` put directly below each atomic process) **in draw.io**. The agent does not edit the drawio directly (it does not rewrite the PFD by any means other than `pfdrenum` / `pfdfix` / `pfdestcallout`).

Before asking for input, state clearly that **these estimates include the review of each deliverable (including addressing review comments)**:

- `楽観` (optimistic): the work volume when nothing goes wrong (unit: business days). Replace the placeholder `d` with something like `2d`.
- `悲観` (pessimistic): the work volume the owner can responsibly commit to (unit: business days). Make it at least the optimistic value.

After input, read it out with `pfdesttable` to check:

```console
$ pfdesttable <project>/pfd.drawio.png
```

Present the output (`ID`, `Description`, `楽観作業量`, `悲観作業量`) to the user to align understanding. While stderr warnings remain (missing box, not filled in, inconsistent values), do not proceed; ask for the corresponding processes to be filled in or corrected (if a box is missing, fill in the scaffolding with `pfdestcallout` in [PFD Fix](#pfd-fix)). Do not ask to change values that are already filled in.

## Master AP Table

The master `<project>/ap.tsv` is where you edit everything **other than** the estimates (needed resources, etc.). **Do not change values the user has already entered** (only blanks get a default at derivation time). Do not include `楽観作業量` / `悲観作業量` columns (the estimate boxes are the source data; for the old format, see the compatibility procedure in [Derive Scenario Tables](#derive-scenario-tables)). The header is determined by the execution model:

- Default (`resource_mode: infinite`, `feedback_mode: disabled`): since neither resources nor rework are handled, the master has no required columns. If it has no columns, you do not need to create the master at all (a table with just `ID` and `説明` (description) is also fine).
  ```tsv
  ID	説明
  ```
- `resource_mode: finite` (consider resource conflicts; add `楽観必要資源` / `悲観必要資源`):
  ```tsv
  ID	説明	楽観必要資源	悲観必要資源
  ```
- `feedback_mode: enabled` (consider rework; add `楽観作業量回復割合`, `悲観作業量回復割合`, and `完了条件`):
  ```tsv
  ID	説明	楽観作業量回復割合	悲観作業量回復割合	完了条件
  ```
- Both (finite resources and rework):
  ```tsv
  ID	説明	楽観作業量回復割合	悲観作業量回復割合	完了条件	楽観必要資源	悲観必要資源
  ```

| Column | Meaning | Default / Notes |
|---|---|---|
| `楽観作業量回復割合` | (With `feedback_mode: enabled`) optimistic rework recovery ratio X (0.0–1.0). With work volume V and n rework cycles, the recovered amount is `V * X^n` | Blank → `0.1` at derivation |
| `悲観作業量回復割合` | (Same as above) pessimistic recovery ratio X (same formula) | Blank → `0.3` at derivation |
| `完了条件` | (Same as above) completion condition of the rework loop. Mapped to pfd-tools' `開始条件` (start condition) at derivation | Blank = `\true` (no constraint). A human fills it in if needed |
| `楽観必要資源` | (With `resource_mode: finite`) optimistic needed resources | Blank → filled at derivation with a default (the optimistic `worker_count` workers) |
| `悲観必要資源` | (Same as above) pessimistic needed resources | Blank → filled at derivation with a default (the pessimistic `worker_count` workers) |

`開始条件`, `グループ` (group), and `見積もり根拠` (estimation basis) are not included in the master (a human adds the columns if needed; all of them are optional columns in pfd-tools). If you use `Description` instead of `説明`, adjust the qhs queries to match the actual header.

### Reading needed resources

Each assignment in `楽観必要資源` / `悲観必要資源` is `ResourceID:work volume consumed per unit time` (the consumed amount may be fractional). Do not confuse the meaning of the separators:

- `;` is **OR** (alternative resources). The planner picks one that is free. Example: `A:1;B:1` means "either A **or** B handles it" (not both).
- `,` is **AND** (simultaneous occupation). All listed resources are used at the same time. Example: `A,B:1` means "occupy both A **and** B".

The default `A:1;B:1;...` (generated by `pfdres`) means "one of that many workers handles it" (OR), i.e. a worker pool shared by all processes. If you do not want to consider resource conflicts, use `resource_mode: infinite` (the `必要資源` column itself becomes unnecessary; this replaces the former approach of "assigning each process a dedicated resource `資源<ID>:1`").

## Derive Scenario Tables

Use `pfdesttable` to read the estimate boxes on the PFD out as a table (`ID`, `Description`, `楽観作業量`, `悲観作業量`), JOIN it with the master `ap.tsv` on `ID`, and `AS` `楽観作業量` / `悲観作業量` to `予想作業量` (expected work volume). The columns you add depend on the execution model:

- With `resource_mode: finite`, `AS` `楽観必要資源` / `悲観必要資源` to `必要資源` (blanks are filled with the default). With `infinite`, do not create the `必要資源` column.
- With `feedback_mode: enabled`, `AS` `作業量回復割合` to `予想手戻り作業量割合` and `完了条件` to `開始条件` (to fill blanks, use `CASE WHEN LENGTH(...) > 0 THEN ... ELSE <default> END`). With `disabled`, do not create `予想手戻り作業量割合`.
- **Always include the `開始条件` column regardless of the execution model.** `pfdplan` / `criticalpath` read this column even with `feedback_mode: disabled`, and fail with `missing precondition column` without it. With `disabled`, put an empty-string literal (`'' AS \"開始条件\"`; blank = `\true` = no constraint).

Processes whose estimates are not filled in (still the placeholder) get an empty `予想作業量`, which the subsequent `pfdlint` reports as an error (unlike bootstrap, it is not filled with `0`; resolve missing entries in the [Estimate Boxes](#estimate-boxes) check before deriving). The JOIN is an inner join on `ID`, so if the rows of the PFD and the master `ap.tsv` are out of sync, rows are silently dropped (missing processes are detected by `pfdlint`).

With the default execution model (`resource_mode: infinite`, `feedback_mode: disabled`), derive only the work volume and an empty `開始条件`. If the master has no columns, no JOIN is needed either; just reshape the `pfdesttable` output:

```console
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT ID, Description AS \"説明\", \"楽観作業量\" AS \"予想作業量\", '' AS \"開始条件\" FROM \"-\"" > <project>/ap_optimistic.tsv
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT ID, Description AS \"説明\", \"悲観作業量\" AS \"予想作業量\", '' AS \"開始条件\" FROM \"-\"" > <project>/ap_pessimistic.tsv
```

With `resource_mode: finite`, add the `必要資源` column. Fill blanks with human resources for that scenario's `worker_count` workers generated by `pfdres <count>` (`pfdres 1` gives `A:1`, `pfdres 2` gives `A:1;B:1`; the one after `Z` is `AA`. This is a worker pool shared by all processes, modeling contention with OR). Worker counts are assumed to be at least 1 in every scenario (0 makes `pfdres` error out). Replace the count in `$(pfdres <worker_count>)` in the following queries with the actual value:

```console
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", t1.\"楽観作業量\" AS \"予想作業量\", CASE WHEN LENGTH(t2.\"楽観必要資源\") > 0 THEN t2.\"楽観必要資源\" ELSE '$(pfdres <optimistic worker_count>)' END AS \"必要資源\", '' AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_optimistic.tsv
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", t1.\"悲観作業量\" AS \"予想作業量\", CASE WHEN LENGTH(t2.\"悲観必要資源\") > 0 THEN t2.\"悲観必要資源\" ELSE '$(pfdres <pessimistic worker_count>)' END AS \"必要資源\", '' AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_pessimistic.tsv
```

With `feedback_mode: enabled`, add `予想手戻り作業量割合` to the queries above and replace the empty `開始条件` with the `AS` of `完了条件` (with `resource_mode: infinite`, just drop the `必要資源` term):

```console
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", t1.\"楽観作業量\" AS \"予想作業量\", CASE WHEN LENGTH(t2.\"楽観作業量回復割合\") > 0 THEN t2.\"楽観作業量回復割合\" ELSE 0.1 END AS \"予想手戻り作業量割合\", CASE WHEN LENGTH(t2.\"楽観必要資源\") > 0 THEN t2.\"楽観必要資源\" ELSE '$(pfdres <optimistic worker_count>)' END AS \"必要資源\", t2.\"完了条件\" AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_optimistic.tsv
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", t1.\"悲観作業量\" AS \"予想作業量\", CASE WHEN LENGTH(t2.\"悲観作業量回復割合\") > 0 THEN t2.\"悲観作業量回復割合\" ELSE 0.3 END AS \"予想手戻り作業量割合\", CASE WHEN LENGTH(t2.\"悲観必要資源\") > 0 THEN t2.\"悲観必要資源\" ELSE '$(pfdres <pessimistic worker_count>)' END AS \"必要資源\", t2.\"完了条件\" AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_pessimistic.tsv
```

Tell the user about the cells filled with defaults and have them fix as needed.

Pipes work as-is in PowerShell too (`pfdesttable <in> | qhs ... > <out>`).

**Compatibility procedure (old-format master)**: this procedure is also a temporary one for compatibility. It targets projects created with skills predating the estimate-box approach (before 6.1.0). **Removal condition**: delete this paragraph once no such projects remain. For old-format projects whose master `ap.tsv` still has `楽観作業量` / `悲観作業量` columns (and whose PFD has no estimate boxes), derive from the master alone as before instead of JOINing. In the queries above, remove `pfdesttable <project>/pfd.drawio.png |`, remove the `t1.` / `t2.` qualifiers, change `FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID` back to `FROM \"-\"`, and pass the master on stdin with `< <project>/ap.tsv`. To migrate to the new format, move the values in the estimate columns into the estimate boxes on the PFD, then delete the columns.

## Project Files And Lint

The scenario settings files (`project_optimistic.json` / `project_pessimistic.json`) were already created by bootstrap, with `atomic_process_table` (and `resource_table` under `resource_mode: finite`) pointing to the per-scenario tables. Since derivation overwrites the same-named tables, no re-creation is needed.

With `resource_mode: infinite`, lint as-is:

```console
$ pfdlint -f <project>/project_optimistic.json
$ pfdlint -f <project>/project_pessimistic.json
```

With `resource_mode: finite`, if you changed resources, regenerate the resource tables (see [Conventions](#conventions)) before linting:

```console
$ pfdtable -t r -res finite -p <project>/pfd.drawio.png -ap <project>/ap_optimistic.tsv -cd <project>/cd.tsv > <project>/r_optimistic.tsv
$ pfdtable -t r -res finite -p <project>/pfd.drawio.png -ap <project>/ap_pessimistic.tsv -cd <project>/cd.tsv > <project>/r_pessimistic.tsv
$ pfdlint -f <project>/project_optimistic.json
$ pfdlint -f <project>/project_pessimistic.json
```

Report the lint results starting from the error / warning counts, listing each issue's severity, target ID, an easy-to-understand explanation of the message, and an improvement suggestion. Do not run `pfdplan` while errors remain. Keep warnings only if the PjM can explain them.

If `feedback-edge-not-available` is reported, the PFD has rework (dashed) edges while `feedback_mode: disabled` is set. Either delete the rework edges, or, if rework should be included in the estimate, set the scenario settings files to `"feedback_mode": "enabled"` and add the columns in [Master AP Table](#master-ap-table).

## Plan Start

Read `start_day` from each scenario settings file.

- If present: present the existing date and confirm whether to use it as-is or a different date (YYYY-MM-DD). On change, update all scenario settings files to the same date.
- If absent (old project): confirm a new entry in YYYY-MM-DD and add it to all scenario settings files.

This date becomes the first task's start date on the Mermaid Gantt / Google Sheets Timeline.

## Generate Plan JSON

Generate the execution plan search result as plan-json only once, and reuse it to render the Gantt and master schedule (the search is heavy, so do not re-search per output format). `-out-format plan-json` overrides `output_format` in the scenario settings file. Business-hours flags (`-start` / `-duration` / `-not-biz-days`) cannot be used for plan-json generation (date assignment is done later by `plantimeline` / `planmaster`). Business-hours settings written in the settings file are simply unused for plan-json, so you can pass the file as-is:

```console
$ pfdplan -f <project>/project_optimistic.json -poor -out-format plan-json > <project>/plan_optimistic.json
$ pfdplan -f <project>/project_pessimistic.json -poor -out-format plan-json > <project>/plan_pessimistic.json
```

If `-poor` does not finish within 20 seconds, fall back to `-poorest` ([Conventions](#quality-fallback)).

## GitHub + Mermaid

When `output_format: mermaid`, render the Gantt (`.mmd`) from the plan-json with `plantimeline`. The output format, start date, and business days are read from the scenario settings file, so do not specify them on the command line:

```console
$ plantimeline -f <project>/project_optimistic.json <project>/plan_optimistic.json > <project>/plan_optimistic.mmd
$ plantimeline -f <project>/project_pessimistic.json <project>/plan_pessimistic.json > <project>/plan_pessimistic.mmd
```

Bootstrap generates the holiday list as `<project>/holidays_ja.txt`, and `not_biz_days` in the scenario settings files refers to it (this does not depend on process substitution, so it also works on Windows / PowerShell). When holidays are updated, regenerate it with `holidays -locale ja > <project>/holidays_ja.txt`.

Emphasize the bars on the critical path. Generate the critical-path TSV with `criticalpath`, and redraw by passing an emphasis ID table narrowed to rows with total float 0 to `-em-tsv` (emphasized bars turn red via mermaid's `crit` tag):

```console
$ criticalpath -f <project>/project_optimistic.json -poor > <project>/criticalpath_optimistic.tsv
$ criticalpath -f <project>/project_pessimistic.json -poor > <project>/criticalpath_pessimistic.tsv
$ qhs -H -O -t -T "SELECT ID FROM <project>/criticalpath_optimistic.tsv WHERE \"最大弾性値（全余裕）\" < 0.0001" > <project>/em_optimistic.tsv
$ qhs -H -O -t -T "SELECT ID FROM <project>/criticalpath_pessimistic.tsv WHERE \"最大弾性値（全余裕）\" < 0.0001" > <project>/em_pessimistic.tsv
$ plantimeline -f <project>/project_optimistic.json -em-tsv <project>/em_optimistic.tsv <project>/plan_optimistic.json > <project>/plan_optimistic.mmd
$ plantimeline -f <project>/project_pessimistic.json -em-tsv <project>/em_pessimistic.tsv <project>/plan_pessimistic.json > <project>/plan_pessimistic.mmd
```

(`最大弾性値（全余裕）` is the maximum elasticity, i.e. total float, column.) For each `criticalpath`, if `-poor` does not finish within 20 seconds, fall back to `-poorest` ([Conventions](#quality-fallback)). Keep `criticalpath_*.tsv`, as it is also used in [Improvement Loop](#improvement-loop).

Keep the `.mmd` as the preview target. Do not create a separate `plan.md`. Optionally, you can convert it to PNG with `mmdc -i <project>/plan_optimistic.mmd -e png` (`mmdc` = mermaid-cli, which requires an extra install; even without it, the `.mmd` can be previewed on GitHub).

## Google Sheets Timeline

When `output_format: google-spreadsheet-tsv`, generate the timeline TSV from the plan-json with `plantimeline`, and the critical-path TSV with `criticalpath`. The output format, start date, and business days are read from the scenario settings file, so do not specify them on the command line:

```console
$ plantimeline -f <project>/project_optimistic.json <project>/plan_optimistic.json > <project>/timeline_optimistic.tsv
$ plantimeline -f <project>/project_pessimistic.json <project>/plan_pessimistic.json > <project>/timeline_pessimistic.tsv
$ criticalpath -f <project>/project_optimistic.json -poor > <project>/criticalpath_optimistic.tsv
$ criticalpath -f <project>/project_pessimistic.json -poor > <project>/criticalpath_pessimistic.tsv
```

For each `criticalpath`, if `-poor` does not finish within 20 seconds, fall back to `-poorest` ([Conventions](#quality-fallback)).

If you pass a table of only the IDs with total float 0 to `plantimeline`'s `-em-tsv`, an `Emphasis` column (`TRUE` for the critical path) is appended to the end of the timeline TSV (same qhs procedure as in [GitHub + Mermaid](#github--mermaid)). If you use this column, the VLOOKUP for the `Critical Path` column below is unnecessary; point the Timeline color and its conditional formatting at the `Emphasis` column instead.

Sheets to create:

| Sheet | Content |
|---|---|
| `Optimistic Timeline` / `Pessimistic Timeline` | Timeline view sourced from `Optimistic Data` / `Pessimistic Data` |
| `Optimistic Data` / `Pessimistic Data` | Paste `timeline_*.tsv` and add the `Critical Path` column (and the `Group` column if `ap.tsv` has a `グループ` column) |
| `Optimistic Critical Path Analysis` / `Pessimistic Critical Path Analysis` | Paste `criticalpath_*.tsv` and add column D `Critical Path` |
| `Atomic Process Table` | Paste `ap.tsv` (the lookup source for `Group` when there is a `グループ` column) |

`グループ` (group) is not a default column of the master `ap.tsv` ([Master AP Table](#master-ap-table)). If you need group display, add a `グループ` column to `ap.tsv` and then use the `Group`-related formulas and settings below. Otherwise, omit the `Group` column, formulas, and the Timeline group setting.

Change the A–H header of `Optimistic Data` / `Pessimistic Data`:

```tsv
Atomic Process	Completions	Allocated Resources	Description	Start Time	End Time	Start	End
```

<details>
<summary>Formulas / conditional formatting / Timeline view settings to add</summary>

I1 of `Optimistic Data` / `Pessimistic Data` (`Group`; only when `ap.tsv` has a `グループ` column. Match the VLOOKUP column number to the position of the `グループ` column in `Atomic Process Table`):

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

Timeline view settings: data range `Optimistic Data!A1:J` / `Pessimistic Data!A1:J`, start date `Start Time`, end date or duration `End Time`, card title `Atomic Process`, color `Critical Path`, details `Description`, group `Group` (only when there is a `Group` column).

Conditional formatting: range `A2:J`, custom formula `=$J2=TRUE`, fill light pink.

</details>

## Master Schedule

(Optional) If you need a top-level Gantt of groups × milestones (a master schedule), generate it from the plan-json with `planmaster`. The input to `planmaster` is a classification table (a TSV of `ID`/`MasterRow`/`MasterBar`); unknown columns are ignored, so you can pass an `ap` that has `MasterRow`/`MasterBar` columns as-is.

Prerequisite: add `MasterRow` (group ID) and `MasterBar` (milestone ID) columns to the master `ap.tsv`, and append `, t2.MasterRow, t2.MasterBar` to the end of the SELECT in [Derive Scenario Tables](#derive-scenario-tables) to carry them into `ap_optimistic.tsv` / `ap_pessimistic.tsv` (`planmaster` errors out without both columns).

`planmaster` reads the output format, start date, and business days from the scenario settings file (`-f`). Always pass the classification table with `-tsv` (use the per-scenario `ap` that has the `MasterRow`/`MasterBar` columns as-is). Passing the `g.tsv` / `m.tsv` prepared in bootstrap's Milestone And Group Tables to `-row-meta` / `-bar-meta` adds descriptions to rows and bars (optional). Pass `-out-format` only when you want an output format different from the settings file (if the settings file's `output_format` is `plan-json` / `timeline-json`, `planmaster` does not accept it, so you must pass it):

```console
$ planmaster -f <project>/project_optimistic.json -tsv <project>/ap_optimistic.tsv -row-meta <project>/g.tsv -bar-meta <project>/m.tsv -out-format mermaid <project>/plan_optimistic.json > <project>/masterschedule_optimistic.mmd
$ planmaster -f <project>/project_pessimistic.json -tsv <project>/ap_pessimistic.tsv -row-meta <project>/g.tsv -bar-meta <project>/m.tsv -out-format mermaid <project>/plan_pessimistic.json > <project>/masterschedule_pessimistic.mmd
```

For pessimistic, always pass `project_pessimistic.json` (which points to `ap_pessimistic.tsv`) (do not reuse the optimistic settings file). Bars depict only the first execution interval of each atomic process (rework is not drawn; the deadline is the first completion). Optionally, you can convert to PNG with `mmdc`.

## Critical Path Review

Start improvement consideration from the optimistic case (because the externally committed master schedule is, as a rule, built from the optimistic plan; if the user specifies a scenario, use that instead). If `criticalpath_optimistic.tsv` does not exist, generate it (reuse it if already generated in [GitHub + Mermaid](#github--mermaid) / [Google Sheets Timeline](#google-sheets-timeline)):

```console
$ criticalpath -f <project>/project_optimistic.json -poor > <project>/criticalpath_optimistic.tsv
```

If `-poor` does not finish within 20 seconds, fall back to `-poorest` ([Conventions](#quality-fallback)).

Explain the processes whose `最大弾性値（全余裕）` (maximum elasticity / total float) is below 0.0001, and the processes whose `最小弾性値` (minimum elasticity) is not `-`.

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

Then re-derive from [Derive Scenario Tables](#derive-scenario-tables); when resources change, regenerate the resource tables (`r_optimistic.tsv` / `r_pessimistic.tsv`) ([Conventions](#conventions)); and finally regenerate the lint, plan-json, Gantt (and master schedule if needed), and critical-path output.

## Replan

If delays occur after the plan is finalized, use `pjm-pfd-replan` to replan (reflecting progress from `progress.tsv`, triaging end-date risk, and presenting the timeline diff).
