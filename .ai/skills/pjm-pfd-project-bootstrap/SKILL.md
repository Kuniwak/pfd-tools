---
name: pjm-pfd-project-bootstrap
description: |
  Bootstrap a PFD project with pfd-tools. Scaffolds the pfd-tools table set from a pfd.drawio.png template, lets you pick the output format (GitHub Mermaid / Google Sheets Timeline), and completes the pre-estimation lint gate. Triggers when the user says things like "start a PFD project", "set up pfd-tools", "prepare for estimation" (or in Japanese, "PFD プロジェクトを始めたい", "pfd-tools をセットアップ", "見積もりの準備").
allowed-tools: Read, Write, Bash, Glob, Grep, readfile, writefile, terminal, glob, grep
---

# pjm-pfd-project-bootstrap

Run this before `pjm-pfd-estimation-plan` to prepare the initial scaffold of a new pfd-tools project and the pre-estimation lint gate.

## Workflow

- [ ] 1. Confirm the working language: ask the user whether to proceed in 日本語 or English. Use the chosen language for all subsequent conversation, explanations, progress reports, and lint / critical-path summaries. (Command names, paths, option flags, column names, and code stay verbatim regardless of the chosen language.)
- [ ] 2. Confirm output format, granularity, worker count, and start date ([Inputs To Confirm](#inputs-to-confirm))
- [ ] 3. Verify pfd-tools / qhs ([Tool Setup](#tool-setup))
- [ ] 4. Copy the scaffold and generate the tables, including the estimate-box scaffolding ([Bootstrap Files](#bootstrap-files))
- [ ] 5. Reshape the master ap.tsv / ad.tsv → derive → lint ([Pre-Estimation Gate](#pre-estimation-gate))
- [ ] 6. Interview the user and fill in the milestone and group tables ([Milestone And Group Tables](#milestone-and-group-tables))
- [ ] 7. Hand off to `pjm-pfd-estimation-plan`

## Conventions

- At the start, confirm the working language (日本語 / English) with the user, then use it consistently for all conversation, explanations, progress reports, and lint summaries. Keep command names, paths, options, column names, and code verbatim.
- Each question in [Inputs To Confirm](#inputs-to-confirm) drives behavior, so always ask for the user's preference (do not decide without context).
- Estimates (optimistic / pessimistic work volumes) are written into the estimate boxes on the PFD (text boxes placed by `pfdestcallout`) and read back as a table with `pfdesttable`. The master `<project>/ap.tsv` is the human-edited source data (it holds columns that cannot live on the PFD, such as needed resources) and is not passed to pfd-tools directly. The pfd-tools inputs are `ap_optimistic.tsv` / `ap_pessimistic.tsv`, derived by JOINing the `pfdesttable` output with the master, plus `project_optimistic.json` / `project_pessimistic.json` (same Design as the follow-up skill).
- Use `project` for `<project>` unless it is clear from context.
- Keep one scenario settings file per scenario (`project_<scenario>.json`). This is not limited to the optimistic / pessimistic pair: multi-point estimates such as three-point estimation are supported simply by adding files.
- The estimation assumptions (output format, execution model, worker count, plan start date) are written into each scenario settings file. `output_format`, `start_day`, and the execution model (`resource_mode` / `feedback_mode`) are keys that pfd-tools itself reads; `scenario` and `worker_count` are keys used only by the skills (pfd-tools ignores unknown keys, so they can coexist).
- The execution model is determined by the estimation granularity and whether there are rework edges. Fine granularity → `resource_mode: finite`; coarse granularity → `infinite` (resource conflicts are not considered, so neither the needed-resources column nor the resource table is required). If the PFD has rework (dashed) edges → `feedback_mode: enabled`; otherwise `disabled`.

## Inputs To Confirm

**1. Execution-plan output format:** (determines `output_format`: GitHub + Mermaid Gantt → `mermaid` / Google Sheets Timeline → `google-spreadsheet-tsv`)

```text
Choose the output format for the execution plan.

1. GitHub + Mermaid Gantt
   - Outputs the pfdplan result as a Mermaid Gantt chart.
   - You can preview the Gantt chart directly from the .mmd file on GitHub.
   - Suited to workflows centered on PR review and diff management on GitHub.

2. Google Sheets Timeline TSV
   - Outputs the pfdplan result as TSV you can paste into Google Sheets.
   - You can review the plan in the Google Sheets Timeline View.
   - Suited to co-editing with a PjM or non-engineers via Google Drive / Sheets.

In either case, the PFD itself is authored in pfd.drawio.png.
pfd.drawio.png previews as an image on GitHub and can be edited with draw.io / diagrams.net on Google Drive.
```

**2. Estimation granularity:** (determines `resource_mode`: fine granularity → `finite` / coarse granularity → `infinite`)

```text
Which estimation granularity do you want?

1. Coarse granularity
  - Each task is sized so that multiple people can share and progress it
  - You grasp the big picture quickly but with low precision. Use this in the early phase of a project
2. Fine granularity
  - Each task is sized so that one person can progress it
  - High-precision estimation is possible, but the details are not clear in the early phase. Use this once the full shape of the project comes into view
```

**3. (Fine granularity only) Worker count:** Ask whether the optimistic and pessimistic scenarios use the same number; if the same, confirm the shared count, otherwise confirm each count (pessimistic must be at or below optimistic, and both at least 1).

```text
Do you want to plan with the same number of workers for the optimistic and pessimistic scenarios?

1. Yes, plan with the same number of workers (e.g., 2 for both optimistic and pessimistic)
2. No, plan with different numbers of workers (e.g., 3 optimistic, 2 pessimistic)
```

**4. Plan start date:**

```text
Specify the start date of the execution plan in YYYY-MM-DD format.
This date becomes the first task's start date on the Mermaid Gantt / Google Sheets Timeline.
```

The confirmed answers are written into the scenario settings files (`<project>/project_optimistic.json` / `<project>/project_pessimistic.json`) in [Bootstrap Files](#bootstrap-files). All keys for one scenario:

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

| Key | Value | Notes |
|---|---|---|
| `pfd`, `atomic_deliverable_table`, `composite_deliverable_table`, `milestone_table`, `group_table` | Path to each file | Shared across all scenarios |
| `atomic_process_table` | That scenario's atomic process table | Differs per scenario (`ap_<scenario>.tsv`) |
| `resource_table` | That scenario's resource table | `resource_mode: finite` only. Differs per scenario (`r_<scenario>.tsv`) |
| `resource_mode` | `finite` (fine granularity) / `infinite` (coarse granularity) | pfd-tools execution model. `infinite` ignores resource conflicts |
| `feedback_mode` | `disabled` (no rework edges) / `enabled` (rework edges present) | pfd-tools execution model |
| `scenario` | Scenario name (`optimistic` / `pessimistic`, etc.) | Add more along with the files for multi-point estimation |
| `output_format` | `mermaid` (GitHub + Mermaid Gantt) / `google-spreadsheet-tsv` (Google Sheets Timeline) | Read by pfd-tools. Use the same value in all scenarios |
| `start_day` | Execution-plan start date (YYYY-MM-DD) | Read by pfd-tools. Use the same value in all scenarios |
| `duration` | Hours per business day | Read by pfd-tools. The command-line default is 9, but the template sets 8 explicitly |
| `not_biz_days` | Path to the holiday list file (relative to the settings file) | Read by pfd-tools. Generated with `holidays` |
| `worker_count` | That scenario's worker count (at least 1) | `resource_mode: finite` only. Pessimistic must be at or below optimistic |

If `<project>/answers.md` remains when re-bootstrapping an existing project, the estimation assumptions are still in the old format. This skill does not read `answers.md`, so tell the user to migrate first with `pjm-pfd-migrate-v7`.

## Tool Setup

Verify that pfd-tools (run `pfdhelp -short` to see the tool descriptions) and `qhs` exist. pfd-tools 7.0.0 or later is required (the version that introduced execution-model selection via `-res` / `-fb` and `resource_mode` / `feedback_mode`, and critical-path emphasis via `-em-tsv`. `pfdestcallout` / `pfdesttable` themselves exist since 6.1.0, but 6.x does not accept this skill's execution-model settings, so guide the user to upgrade).

If not found, install:

- **pfd-tools**: download the latest binary from <https://github.com/Kuniwak/pfd-tools/releases> and place it on a directory in your PATH (do not edit PATH without approval).
- **qhs** (macOS/Linux): `brew tap itchyny/tap && brew install qhs`. Otherwise use the GitHub Releases binary (`gh release view/download --repo itchyny/qhs`; if `gh` is unavailable, `curl` / PowerShell `Invoke-WebRequest`), and for unsupported OS/arch, `stack install qhs`. Do not hardcode asset names; check the latest.

## Bootstrap Files

For GitHub + Mermaid Gantt only, verify you are inside a git repository (`git rev-parse --show-toplevel`), and if outside, run `git init` after approval.

Create the project directory and copy the scaffold:

```console
$ mkdir -p <project>
```

The scaffold `assets/` contains `basic-pfd.drawio.png` (PFD template) and `project_optimistic.json` / `project_pessimistic.json` (scenario settings templates). First, copy `basic-pfd.drawio.png` to `<project>/pfd.drawio.png` (absolute paths are environment-dependent):

- Claude Code: `cp ${CLAUDE_PLUGIN_ROOT}/skills/pjm-pfd-project-bootstrap/assets/basic-pfd.drawio.png <project>/pfd.drawio.png`
- `${CLAUDE_PLUGIN_ROOT}` undefined (Cursor, etc.): normalize the absolute path of this SKILL.md with `dirname` and `cp` the `assets/...` at the same level. If it cannot be resolved, do a limited search: `find ~/.cursor ~/.config/cursor ~/.agents ~/.claude -name basic-pfd.drawio.png -path '*pjm-pfd-project-bootstrap*' 2>/dev/null | head -1`

Then, in the same way, copy `project_optimistic.json` / `project_pessimistic.json` from `assets/` into `<project>/` and write in the answers from [Inputs To Confirm](#inputs-to-confirm) (if they already exist, confirm this is a re-bootstrap of the same project before overwriting). The templates assume `resource_mode: finite`, so for coarse granularity (`infinite`) change them as follows:

- Set `resource_mode` to `infinite`
- Delete the `resource_table` and `worker_count` keys (the infinite-resource model does not read resources)

Write the same values of `output_format`, `start_day`, and `duration` into both files, and the per-scenario worker count into `worker_count`. Also generate the holiday list (referenced by `not_biz_days`; it avoids process substitution, so it also works on Windows / PowerShell):

```console
$ holidays -locale ja > <project>/holidays_ja.txt
```

Number the IDs with `pfdrenum`, connect "edges that look attached to a rectangle but have no `source`/`target` set" with `pfdfix`, add an estimate box (placeholder `楽観: d 悲観: d`, i.e. "optimistic: d pessimistic: d") directly below each atomic process with `pfdestcallout`, and then generate the tables. For each of `pfdrenum`, `pfdfix`, and `pfdestcallout`, check the stdout result without `-inplace` first, then apply `-inplace`:

```console
$ pfdrenum <project>/pfd.drawio.png
$ pfdrenum -inplace <project>/pfd.drawio.png
$ pfdfix <project>/pfd.drawio.png
$ pfdfix -inplace <project>/pfd.drawio.png
$ pfdestcallout <project>/pfd.drawio.png
$ pfdestcallout -inplace <project>/pfd.drawio.png
$ pfdtable -t all-plan -res <finite|infinite> -fb <disabled|enabled> -p <project>/pfd.drawio.png -out-dir <project>
```

Align `-res` / `-fb` with the execution model written in the scenario settings. `pfdtable` does not generate columns or tables that the execution model does not read: with `infinite`, the `必要資源` (needed resources) column and `r.tsv` are not created; with `disabled`, the `予想手戻り作業量割合` (expected rework volume ratio) and `最大版` (max version) columns are not created. The `<project>/project.json` generated by `pfdtable` is replaced by the scenario settings and is not used (it is deleted below).

The estimate boxes are placed as `text` shapes on the same layer as the processes and do not affect the judgments of `pfdlint` / `pfdplan` / `pfdtable` (they are not recognized as processes or deliverables). A human fills them in on draw.io during the estimation phase (`pjm-pfd-estimation-plan`), e.g. `楽観: 2d 悲観: 3d` (`d` is in business days). `pfdestcallout` is idempotent (it leaves atomic processes that already have a box directly below unchanged), so after adding processes to the PFD you can rerun it to fill in the scaffolding.

By default, `pfdfix` does not delete edges whose endpoint falls inside no rectangle (unhit) or that would become self-loops; it keeps them and emits a WARN. **Do not pass `-delete-unhit` unless the user explicitly asks.** If a WARN appears, tell the user that the connection target is ambiguous or unconnected and let them decide whether to revise the drawing on the PFD. Because `pfdrenum` / `pfdfix` change the graph's connections, always confirm with `pfdlint` in the subsequent [Pre-Estimation Gate](#pre-estimation-gate) that no new soundness violations (cycles, single-src violations, etc.) have appeared.

For GitHub + Mermaid Gantt, commit the scaffold baseline (it makes the subsequent diff easier to read; do not use `git add .`, and if there are unrelated changes under `<project>`, confirm them first):

```console
$ git status --short -- <project>
$ git add <project>
$ git commit -m "Bootstrap PFD project"
```

## Pre-Estimation Gate

Before handing off to estimation, reshape the generated `<project>/ap.tsv` and `<project>/ad.tsv` into the master format ([Conventions](#conventions): the master `ap.tsv` is not passed to pfd-tools directly; `ad.tsv` is kept in a form pfd-tools can read as-is).

### ad.tsv (deliverable table) to master format

Reconstruct the columns in a TSV-aware way (no whole-file regex replacement). Rename `Description` to `説明` (description). Use the following header (`フォーマット` (format), `レビュー基準` (review criteria), `レビューア` (reviewer), and `URL` are left blank; pfd-tools ignores unknown columns, so `ad.tsv` can be used as input as-is):

```tsv
ID	説明	利用可能時刻	フォーマット	レビュー基準	レビューア	URL
```

With `feedback_mode: enabled` (rework edges present), add a `最大版` (max version) column after `利用可能時刻` (available time) and fill in the `最大版` (version cap, integer ≥ 1) of each feedback-source deliverable to terminate the loop (non-feedback deliverables are blank or `-`). With `feedback_mode: disabled`, `最大版` is not read, so the column is unnecessary. `利用可能時刻` only takes effect for initial deliverables with no upstream process (blank/0 means "available from the start").

### ap.tsv (atomic process table) to master format

First check whether the PFD has **rework (dashed) edges** (see "feedback edges = dashed arrows" in the README). If it does, change the scenario settings to `feedback_mode: enabled` (otherwise keep the default `disabled`). Reconstruct the columns in a TSV-aware way and drop the pfdtable-generated `予想作業量` (expected work volume) (and `必要資源`) (keep `ID`, rename `Description` to `説明`; do not include `開始条件` (start condition), `グループ` (group), or `見積もり根拠` (estimation basis) in the master). **Do not put `楽観作業量` / `悲観作業量` (optimistic / pessimistic work volume) columns in the master** (estimates are written into the estimate boxes on the PFD and read by `pfdesttable` at derivation time). Choose the header by execution model:

- Default (`resource_mode: infinite`, `feedback_mode: disabled`): no columns are required in the master.
  ```tsv
  ID	説明
  ```
- `resource_mode: finite` (adds `楽観必要資源` / `悲観必要資源`, optimistic / pessimistic needed resources):
  ```tsv
  ID	説明	楽観必要資源	悲観必要資源
  ```
- `feedback_mode: enabled` (adds `楽観作業量回復割合` / `悲観作業量回復割合`, optimistic / pessimistic rework volume ratio, and `完了条件`, completion condition):
  ```tsv
  ID	説明	楽観作業量回復割合	悲観作業量回復割合	完了条件
  ```
- Both: has all of the above columns.
  ```tsv
  ID	説明	楽観作業量回復割合	悲観作業量回復割合	完了条件	楽観必要資源	悲観必要資源
  ```

1. Estimates go into the estimate boxes on the PFD (they may stay as the placeholder `楽観: d 悲観: d` until the estimation phase; blanks are filled with `0` at derivation time and `pfdesttable` warns about them).
2. With `feedback_mode: enabled`, `楽観作業量回復割合` / `悲観作業量回復割合` may be left blank (optimistic `0.1` / pessimistic `0.3` are filled in at derivation time). Leave `完了条件` blank (blank = `\true` = no constraint; a human can fill it in later if needed).
3. With `resource_mode: finite`, fill in `楽観必要資源` / `悲観必要資源`. Use values like `A:1`, assigning that scenario's `worker_count` workers to each column (if blank, the `pfdres` default is filled in). Tell the user which cells were filled in and prompt them to fix them.

The default value filled into a blank `必要資源` is that scenario's `worker_count` human resources, generated with `pfdres <count>` (`pfdres 1` gives `A:1`, `pfdres 2` gives `A:1;B:1`; the one after `Z` is `AA`. It models contention as an OR over a pool shared by all processes). Worker counts are assumed to be at least 1 in every scenario (`pfdres` errors on 0). If you do not want to consider resource conflicts, use `resource_mode: infinite` (the `必要資源` column itself becomes unnecessary).

**How to read needed resources**: Each allocation is `ResourceID:work-volume-consumed-per-unit-time` (the consumption may be fractional). `;` means **OR** (alternative resources; pick one. E.g. `A:1;B:1` means "one person, A **or** B"), and `,` means **AND** (occupied simultaneously. E.g. `A,B:1` means "occupy both"). The default `A:1;B:1;...` (generated by `pfdres`) means "any one of that many people", i.e. an OR over a shared pool.

### Deriving and linting the optimistic / pessimistic scenarios

Read the estimate boxes on the PFD as a table (`ID`, `Description`, `楽観作業量`, `悲観作業量`) with `pfdesttable`, JOIN it with the master `ap.tsv` on `ID`, and derive each scenario. Before estimation (still placeholders), `楽観作業量` / `悲観作業量` are empty, so fill them with `0` to pass the gate (report the `pfdesttable` warnings as the list of estimates not yet filled in). The columns to add depend on the execution model (`必要資源` for `resource_mode: finite`, `予想手戻り作業量割合` for `feedback_mode: enabled`).

**Always include the `開始条件` (start condition) column regardless of execution model.** `pfdplan` / `criticalpath` read this column even with `feedback_mode: disabled`, and fail with `missing precondition column` if it is absent. With `feedback_mode: disabled`, use an empty string literal (`'' AS \"開始条件\"`; blank = `\true` = no constraint); with `enabled`, alias the master's `完了条件`.

For the default (`resource_mode: infinite`, `feedback_mode: disabled`), if the master has no extra columns, no JOIN is needed either (`開始条件` is added as an empty column):

```console
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT ID, Description AS \"説明\", CASE WHEN LENGTH(\"楽観作業量\") > 0 THEN \"楽観作業量\" ELSE 0 END AS \"予想作業量\", '' AS \"開始条件\" FROM \"-\"" > <project>/ap_optimistic.tsv
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT ID, Description AS \"説明\", CASE WHEN LENGTH(\"悲観作業量\") > 0 THEN \"悲観作業量\" ELSE 0 END AS \"予想作業量\", '' AS \"開始条件\" FROM \"-\"" > <project>/ap_pessimistic.tsv
```

`resource_mode: finite` adds the `必要資源` column (replace the count in `$(pfdres <worker_count>)` in the queries below with the actual value):

```console
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", CASE WHEN LENGTH(t1.\"楽観作業量\") > 0 THEN t1.\"楽観作業量\" ELSE 0 END AS \"予想作業量\", CASE WHEN LENGTH(t2.\"楽観必要資源\") > 0 THEN t2.\"楽観必要資源\" ELSE '$(pfdres <optimistic worker_count>)' END AS \"必要資源\", '' AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_optimistic.tsv
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", CASE WHEN LENGTH(t1.\"悲観作業量\") > 0 THEN t1.\"悲観作業量\" ELSE 0 END AS \"予想作業量\", CASE WHEN LENGTH(t2.\"悲観必要資源\") > 0 THEN t2.\"悲観必要資源\" ELSE '$(pfdres <pessimistic worker_count>)' END AS \"必要資源\", '' AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_pessimistic.tsv
```

`feedback_mode: enabled` adds `予想手戻り作業量割合` to the queries above (filling blanks in `作業量回復割合`) and replaces the empty `開始条件` with an alias of `完了条件` (blank = `\true`) (with `resource_mode: infinite`, just remove the `必要資源` term):

```console
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", CASE WHEN LENGTH(t1.\"楽観作業量\") > 0 THEN t1.\"楽観作業量\" ELSE 0 END AS \"予想作業量\", CASE WHEN LENGTH(t2.\"楽観作業量回復割合\") > 0 THEN t2.\"楽観作業量回復割合\" ELSE 0.1 END AS \"予想手戻り作業量割合\", CASE WHEN LENGTH(t2.\"楽観必要資源\") > 0 THEN t2.\"楽観必要資源\" ELSE '$(pfdres <optimistic worker_count>)' END AS \"必要資源\", t2.\"完了条件\" AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_optimistic.tsv
$ pfdesttable <project>/pfd.drawio.png | qhs -H -O -t -T "SELECT t1.ID, t2.\"説明\", CASE WHEN LENGTH(t1.\"悲観作業量\") > 0 THEN t1.\"悲観作業量\" ELSE 0 END AS \"予想作業量\", CASE WHEN LENGTH(t2.\"悲観作業量回復割合\") > 0 THEN t2.\"悲観作業量回復割合\" ELSE 0.3 END AS \"予想手戻り作業量割合\", CASE WHEN LENGTH(t2.\"悲観必要資源\") > 0 THEN t2.\"悲観必要資源\" ELSE '$(pfdres <pessimistic worker_count>)' END AS \"必要資源\", t2.\"完了条件\" AS \"開始条件\" FROM \"-\" AS t1 JOIN <project>/ap.tsv AS t2 ON t1.ID = t2.ID" > <project>/ap_pessimistic.tsv
```

The JOIN is an inner join on `ID`, so rows are silently dropped if the PFD and the master `ap.tsv` are out of sync. `pfdlint` after derivation detects missing processes, so always run lint.

With `resource_mode: finite`, next generate each scenario's resource table (not needed for `infinite`):

```console
$ pfdtable -t r -res finite -p <project>/pfd.drawio.png -ap <project>/ap_optimistic.tsv -cd <project>/cd.tsv > <project>/r_optimistic.tsv
$ pfdtable -t r -res finite -p <project>/pfd.drawio.png -ap <project>/ap_pessimistic.tsv -cd <project>/cd.tsv > <project>/r_pessimistic.tsv
```

(The resource tables are split per scenario because the resource ID set differs when the scenarios have different worker counts.)

Delete the unused `project.json` generated by `pfdtable` (with `resource_mode: finite`, also delete the unused `r.tsv`, since the per-scenario resource tables replace it). Then lint each scenario:

```console
$ rm <project>/project.json <project>/r.tsv
$ pfdlint -f <project>/project_optimistic.json
$ pfdlint -f <project>/project_pessimistic.json
```

Report the lint results starting from the error / warning counts, listing each issue's severity, target ID, message, and a fix candidate. Do not proceed to estimation while errors remain. Keep warnings only if the PjM can explain them.

## Milestone And Group Tables

Fill in `<project>/g.tsv` (`ID`/`Description`) and `<project>/m.tsv` (`ID`/`Description`/`Groups`/`Successors`) generated by `pfdtable -t all-plan`, interviewing the user one question at a time (edit them in a TSV-aware way). These are used as display descriptions (`-row-meta` / `-bar-meta`) when building a master schedule with `planmaster` (the current `pfdplan` / `criticalpath` do not consume them; if you will not build a master schedule, they may stay empty).

### Group table g.tsv (choose granularity by scale)

First confirm the scale of the project and decide the group granularity:

- **Large**: split by functional domain (e.g., onboarding / order intake / reports / infrastructure, one group per domain). Have the user list the domains and write one row per domain in `g.tsv`.
- **Small to medium**: consolidate into a single group (`G1` only).

The master schedule's classification table is given by the `MasterRow` and `MasterBar` columns in `ap.tsv` (`MasterRow` = group ID, `MasterBar` = milestone ID; neither is a default column). With these two columns, `ap.tsv` can be passed to `planmaster -tsv` as-is.

### Milestone table m.tsv (Successors and overlap)

Interview the user for each milestone's `ID`, `Description`, and the `Groups` it belongs to (comma-separated group IDs), and write them in. Decide `Successors` (comma-separated IDs of succeeding milestones) as follows:

- **Empty by default** is fine. `planmaster` does not read `Successors` (bars are computed from each atomic process's first-execution interval, so ordering declarations do not affect rendering).
- You may fill in `Successors` as a note recording the intended milestone order and for consistency checking (`pfdlint`).

When done, tell the user that the next step is to edit the PFD and then use `pjm-pfd-estimation-plan`.
