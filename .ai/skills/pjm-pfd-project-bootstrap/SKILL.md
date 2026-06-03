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
- [ ] 2. Confirm output format, granularity, worker count, and start date, and save to `answers.md` ([Inputs To Confirm](#inputs-to-confirm))
- [ ] 3. Verify pfd-tools / qhs ([Tool Setup](#tool-setup))
- [ ] 4. Copy the scaffold and generate tables ([Bootstrap Files](#bootstrap-files))
- [ ] 5. Reshape the master ap.tsv → derive → create project_* → lint ([Pre-Estimation Gate](#pre-estimation-gate))
- [ ] 6. Hand off to `pjm-pfd-estimation-plan`

## Conventions

- At the start, confirm the working language (日本語 / English) with the user, then use it consistently for all conversation, explanations, progress reports, and lint summaries. Keep command names, paths, options, column names, and code verbatim.
- Each question in [Inputs To Confirm](#inputs-to-confirm) drives behavior, so always ask for the user's preference (do not decide without context).
- The master `<project>/ap.tsv` is the human-edited source data and is not passed to pfd-tools. The pfd-tools inputs are the derived `ap_optimistic.tsv` / `ap_pessimistic.tsv` and `project_optimistic.json` / `project_pessimistic.json` (same Design as the follow-up skill).
- Use `project` for `<project>` unless it is clear from context.

## Inputs To Confirm

**1. Execution-plan output format:**

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

**2. Estimation granularity:**

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

After confirmation, save to `<project>/answers.md`:

```yaml
project: <project>
output_format: github-mermaid
resource_conflicts: enabled
optimistic_worker_count: 2
pessimistic_worker_count: 2
plan_start_date: 2026-05-20
```

| Key | Value | Notes |
|---|---|---|
| `output_format` | `github-mermaid` / `google-sheets-timeline` | Output format |
| `resource_conflicts` | `enabled` (fine granularity) / `disabled` (coarse granularity) | `disabled` triggers Deconfliction downstream |
| `optimistic_worker_count` | Optimistic worker count (at least 1) | Filled in when `enabled` (fine granularity); omitted when `disabled` |
| `pessimistic_worker_count` | Pessimistic worker count (at least 1) | Filled in when `enabled` (fine granularity); omitted when `disabled` |
| `plan_start_date` | Execution-plan start date | YYYY-MM-DD |

## Tool Setup

Verify that pfd-tools (`pfdrenum` `pfdtable` `pfdlint` `pfdplan` `criticalpath` `holidays`) and `qhs` exist via `-h` / `--help`. pfd-tools 5.1.0 or later is required.

If not found, install:

- **pfd-tools**: download the latest binary from <https://github.com/Kuniwak/pfd-tools/releases> and place it on a directory in your PATH (do not edit PATH without approval).
- **qhs** (macOS/Linux): `brew tap itchyny/tap && brew install qhs`. Otherwise use the GitHub Releases binary (`gh release view/download --repo itchyny/qhs`; if `gh` is unavailable, `curl` / PowerShell `Invoke-WebRequest`), and for unsupported OS/arch, `stack install qhs`. Do not hardcode asset names; check the latest.

## Bootstrap Files

For GitHub + Mermaid Gantt only, verify you are inside a git repository (`git rev-parse --show-toplevel`), and if outside, run `git init` after approval.

Create the project directory and copy the scaffold:

```console
$ mkdir -p <project>
```

The scaffold `assets/` contains `basic-pfd.drawio.png` (PFD template) and `project_optimistic.json` / `project_pessimistic.json` (per-scenario project settings). First, copy `basic-pfd.drawio.png` to `<project>/pfd.drawio.png` (the project settings are copied in [Pre-Estimation Gate](#pre-estimation-gate); absolute paths are environment-dependent):

- Claude Code: `cp ${CLAUDE_PLUGIN_ROOT}/skills/pjm-pfd-project-bootstrap/assets/basic-pfd.drawio.png <project>/pfd.drawio.png`
- `${CLAUDE_PLUGIN_ROOT}` undefined (Cursor, etc.): normalize the absolute path of this SKILL.md with `dirname` and `cp` the `assets/...` at the same level. If it cannot be resolved, do a limited search: `find ~/.cursor ~/.config/cursor ~/.agents ~/.claude -name basic-pfd.drawio.png -path '*pjm-pfd-project-bootstrap*' 2>/dev/null | head -1`

Write `<project>/answers.md` immediately after creating it (if it already exists, confirm whether this is a re-bootstrap of the same project before overwriting).

Apply `pfdrenum` safely (check stdout → `-inplace`) and generate the tables:

```console
$ pfdrenum <project>/pfd.drawio.png
$ pfdrenum -inplace <project>/pfd.drawio.png
$ pfdtable -t all-plan -locale en -p <project>/pfd.drawio.png -out-dir <project>
```

For GitHub + Mermaid Gantt, commit the scaffold baseline (it makes the subsequent diff easier to read; do not use `git add .`, and if there are unrelated changes under `<project>`, confirm them first):

```console
$ git status --short -- <project>
$ git add <project>
$ git commit -m "Bootstrap PFD project"
```

## Pre-Estimation Gate

Before handing off to estimation, reshape the generated `<project>/ap.tsv` into the master format ([Conventions](#conventions): the master is not passed to pfd-tools).

1. Reconstruct the columns in a TSV-aware way. Drop the pfdtable-generated `Est. Work Volume`, `Est. Rework Volume Ratio`, and `Needed Resources` (keep `ID`, `Description`, `Start Condition`) and use this header:

   ```tsv
   ID	Description	Optimistic Work Volume	Pessimistic Work Volume	Optimistic Rework Volume Ratio	Pessimistic Rework Volume Ratio	Optimistic Needed Resources	Pessimistic Needed Resources	Start Condition	Group	Estimation Basis
   ```

2. Before estimation, set `Optimistic Work Volume` / `Pessimistic Work Volume` to `0` (the unit is business days). The rework ratios may be left blank (the follow-up fills in optimistic `0.1` / pessimistic `0.3`).
3. Fill in `Optimistic Needed Resources` / `Pessimistic Needed Resources`. For coarse granularity, use the same value in both columns, e.g. `DeptA:1`. For fine granularity, use `WorkerA:1` etc., assigning `optimistic_worker_count` / `pessimistic_worker_count` workers to each column. They may be blank (a default is filled in at derivation time, and the filled cells are reported to the user with a prompt to fix them).

The default value filled into a blank `Needed Resources` is, for fine granularity (`enabled`), `optimistic_worker_count` / `pessimistic_worker_count` workers as `WorkerA:1;WorkerB:1;...;WorkerN:1` (`1` → `WorkerA:1`, `2` → `WorkerA:1;WorkerB:1`; the one after `WorkerZ` is `WorkerAA`), and for coarse granularity (`disabled`), `DeptA:1`. Worker counts are assumed to be at least 1 for both optimistic and pessimistic (a count of 0 would make the default empty).

Derive the optimistic and pessimistic tables, generate each scenario's resource table, and prepare the project settings files (replace `WorkerA:1;WorkerB:1;...;WorkerN:1` / `WorkerA:1;WorkerB:1;...;WorkerM:1` in the queries below with the default values above):

```console
$ qhs -H -O -t -T "SELECT ID, Description, \"Optimistic Work Volume\" AS \"Est. Work Volume\", CASE WHEN LENGTH(\"Optimistic Rework Volume Ratio\") > 0 THEN \"Optimistic Rework Volume Ratio\" ELSE 0.1 END AS \"Est. Rework Volume Ratio\", CASE WHEN LENGTH(\"Optimistic Needed Resources\") > 0 THEN \"Optimistic Needed Resources\" ELSE 'WorkerA:1;WorkerB:1;...;WorkerN:1' END AS \"Needed Resources\", \"Start Condition\", \"Group\", \"Estimation Basis\" FROM \"-\"" < <project>/ap.tsv > <project>/ap_optimistic.tsv
$ qhs -H -O -t -T "SELECT ID, Description, \"Pessimistic Work Volume\" AS \"Est. Work Volume\", CASE WHEN LENGTH(\"Pessimistic Rework Volume Ratio\") > 0 THEN \"Pessimistic Rework Volume Ratio\" ELSE 0.3 END AS \"Est. Rework Volume Ratio\", CASE WHEN LENGTH(\"Pessimistic Needed Resources\") > 0 THEN \"Pessimistic Needed Resources\" ELSE 'WorkerA:1;WorkerB:1;...;WorkerM:1' END AS \"Needed Resources\", \"Start Condition\", \"Group\", \"Estimation Basis\" FROM \"-\"" < <project>/ap.tsv > <project>/ap_pessimistic.tsv
$ pfdtable -t r -locale en -p <project>/pfd.drawio.png -ap <project>/ap_optimistic.tsv -cd <project>/cd.tsv > <project>/r_optimistic.tsv
$ pfdtable -t r -locale en -p <project>/pfd.drawio.png -ap <project>/ap_pessimistic.tsv -cd <project>/cd.tsv > <project>/r_pessimistic.tsv
```

`project_optimistic.json` / `project_pessimistic.json` are already provided in the scaffold `assets/`, with `atomic_process_table` / `resource_table` pointing to the per-scenario tables (`ap_optimistic.tsv` + `r_optimistic.tsv` / `ap_pessimistic.tsv` + `r_pessimistic.tsv`) (the resource tables are split because the resource ID set differs when the optimistic and pessimistic worker counts differ). Resolve `assets/` the same way as in [Bootstrap Files](#bootstrap-files) and copy both files into `<project>/`. Delete the unused `project.json` and `r.tsv` that pfdtable generated, then lint both scenarios:

```console
$ rm <project>/project.json <project>/r.tsv
$ pfdlint -f <project>/project_optimistic.json
$ pfdlint -f <project>/project_pessimistic.json
```

Report the lint results starting from the error / warning counts, listing each issue's severity, target ID, message, and a fix candidate. Do not proceed to estimation while errors remain. Keep warnings only if the PjM can explain them. When done, tell the user that the next step is to edit the PFD and then use `pjm-pfd-estimation-plan`.
