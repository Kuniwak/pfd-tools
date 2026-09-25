---
name: pjm-pfd-migrate-v7
description: |
  Migrate a PFD project created with the pfd-tools 6.6.x-or-earlier skills to the 7.0.0-or-later format. Moves the estimation assumptions from `answers.md` into the scenario settings files (project_<scenario>.json), fills in settings that only the old format lacked, and deletes `answers.md`. Triggers when the user says things like "there is an answers.md", "I want to migrate an old-format project", "after upgrading pfd-tools I can't create a plan / settings don't take effect" (or in Japanese, "answers.md がある", "古い形式のプロジェクトを移行したい", "pfd-tools を上げたら計画が作れない・設定が効かない"), or when another skill detects `<project>/answers.md`. Does not run estimation or planning itself (hands off to pjm-pfd-estimation-plan after migration).
allowed-tools: Read, Write, Bash, Glob, Grep, readfile, writefile, terminal, glob, grep
---

# pjm-pfd-migrate-v7

Migrate an old-format PFD project that has `answers.md` to the 7.0.0-or-later scenario settings file format. **Knowledge of `answers.md` lives only in this skill.** Other skills do not know the contents, keys, or conversion rules of `answers.md`; they only check for the file's existence and delegate to this skill.

## Scope

| Item | Details |
|---|---|
| Source | Projects created with the bootstrap that kept estimation assumptions in `answers.md` (pfd-tools **6.6.x or earlier**) |
| Target | The format consolidated into scenario settings files `project_<scenario>.json` (**7.0.0 or later**) |
| Detection | If `<project>/answers.md` exists, the project is a migration target. Otherwise this skill does nothing |
| Out of scope | Running estimation and planning (`pjm-pfd-estimation-plan`), migrating from the old estimate columns in the master `ap.tsv` (the compatibility procedure in `pjm-pfd-estimation-plan`) |

## Naming Convention (for multi-step migrations)

Create **one migration skill per target major version**, named `pjm-pfd-migrate-v<target major version>`. For projects spanning multiple versions, apply them in **ascending** version order (e.g., to bring a 6.x project to 9.x: `pjm-pfd-migrate-v7` → `pjm-pfd-migrate-v8` → `pjm-pfd-migrate-v9`). Each skill handles only one hop and has no knowledge of the hops before or after it.

When adding a new migration skill, **confine the knowledge of the settings, files, and columns removed in that hop to that skill**. Do not write anything in other skills beyond "if that old file exists, invoke the corresponding migration skill".

## Removal Criteria

Once no projects need migration, **delete this skill's entire directory**. Also remove the delegation notes in these two places:

- The delegation on detecting `answers.md` in the Scenario Settings of `pjm-pfd-estimation-plan`
- The delegation for re-bootstrapping at the end of Inputs To Confirm in `pjm-pfd-project-bootstrap`

This is a compatibility path, so **do not grow it**. When keys are added to the target format, first consider whether this skill can be removed before extending [Key Mapping](#key-mapping).

## Workflow

- [ ] 1. Confirm the working language: ask the user whether to proceed in 日本語 or English. Use the chosen language for all subsequent conversation, explanations, progress reports, and lint / critical-path summaries. (Command names, paths, option flags, column names, and code stay verbatim regardless of the chosen language.)
- [ ] 2. Check that `<project>/answers.md` exists and inspect its contents ([Detect](#detect))
- [ ] 3. Present the migration to the user and get agreement ([Detect](#detect))
- [ ] 4. Move the keys into the scenario settings files ([Key Mapping](#key-mapping))
- [ ] 5. Fill in settings not present in `answers.md` ([Missing Settings](#missing-settings))
- [ ] 6. Clean up according to `resource_mode` ([Cleanup By Resource Mode](#cleanup-by-resource-mode))
- [ ] 7. Confirm with `pfdlint` that the migrated inputs are valid ([Verify](#verify))
- [ ] 8. Delete `answers.md` and hand off to `pjm-pfd-estimation-plan` ([Finish](#finish))

## Conventions

- At the start, confirm the working language (日本語 / English) with the user, then use it consistently for all conversation, explanations, and progress reports. Keep command names, paths, options, key names, and column names verbatim.
- Migrate only **after getting the user's agreement**. Old values can be traced in git history, but they are lost if the project is not under version control, so check whether it is.
- If the scenario settings files already exist, **do not overwrite keys that are already present**. Add only the missing keys.
- Do not touch the estimates (work volumes) or the PFD. The only things this skill changes are the scenario settings files, `answers.md`, and (when migrating to `resource_mode: infinite`) the resource tables and the master's resource columns.

## Detect

If `<project>/answers.md` exists, the project is a migration target. The old format has the following YAML block.

````markdown
# pjm-pfd-project-bootstrap answers

```yaml
project: sample
output_format: github-mermaid
resource_conflicts: enabled
optimistic_worker_count: 3
pessimistic_worker_count: 3
plan_start_date: 2026-06-01
```
````

Present to the user what you read and what will be written where according to [Key Mapping](#key-mapping) / [Missing Settings](#missing-settings), and proceed after getting agreement. If the scenario settings files (`project_optimistic.json` / `project_pessimistic.json`) do not exist, create them using the full set of keys for one scenario in `pjm-pfd-project-bootstrap` as a template.

## Key Mapping

| Key in `answers.md` | Destination | Conversion |
|---|---|---|
| `output_format: github-mermaid` | `output_format` in each `project_*.json` | `mermaid` |
| `output_format: google-sheets-timeline` | Same as above | `google-spreadsheet-tsv` |
| `resource_conflicts: enabled` | `resource_mode` in each `project_*.json` | `finite` |
| `resource_conflicts: disabled` | Same as above | `infinite` |
| `optimistic_worker_count` | `worker_count` in `project_optimistic.json` | As-is (not needed with `resource_mode: infinite`, so do not write it) |
| `pessimistic_worker_count` | `worker_count` in `project_pessimistic.json` | Same as above |
| `plan_start_date` | `start_day` in each `project_*.json` | As-is |
| `project` | No destination | Discard; the project directory name suffices |

Use the same `output_format` and `start_day` values in all scenarios. Write the per-scenario worker count into `worker_count`.

## Missing Settings

The following keys are **absent from `answers.md` because the old format passed them on the command line**. The current format reads them from the scenario settings files, so always add them during migration. If you omit them, the plan is computed with default values **without any warning**.

| Destination | How to decide | What happens if omitted |
|---|---|---|
| `duration` in each `project_*.json` | Hours per business day. Set a value that fits your practice (e.g. `8`) explicitly | Falls back to the command-line default `9` |
| `not_biz_days` in each `project_*.json` | Create the holiday list with `holidays -locale ja > <project>/holidays_ja.txt` and write its path (relative to the settings file) | Holidays are treated as business days |
| `feedback_mode` in each `project_*.json` | Decide by whether the PFD has rework (dashed) edges: `enabled` if it does, `disabled` otherwise | Falls back to the pfd-tools default `disabled`. If there are rework edges, `pfdlint` reports `feedback-edge-not-available` |
| `scenario` in each `project_*.json` | That scenario's name (`optimistic` / `pessimistic`) | A skill-only key, so it does not affect pfd-tools' computation |

Generate the holiday list:

```console
$ holidays -locale ja > <project>/holidays_ja.txt
```

## Cleanup By Resource Mode

When migrating from `resource_conflicts: disabled` to `resource_mode: infinite`, the model no longer reads resources, so the following become unnecessary. **Confirm deletion with the user.**

- The `resource_table` and `worker_count` keys in each `project_*.json`
- `<project>/r_optimistic.tsv` / `r_pessimistic.tsv` (and the old `r.tsv`)
- The `楽観必要資源` / `悲観必要資源` (optimistic / pessimistic needed resources) columns in the master `<project>/ap.tsv`
- The `必要資源` (needed resources) column in the derived tables `ap_*.tsv` (it disappears on the next derivation, and `pfdlint` passes even if left as-is)

When migrating from `resource_conflicts: enabled` to `resource_mode: finite`, keep the resource tables and the master's resource columns as they are.

## Verify

Lint with the migrated scenario settings files. If the derived tables are stale, re-derive them with Derive Scenario Tables in `pjm-pfd-estimation-plan` before linting.

```console
$ pfdlint -f <project>/project_optimistic.json
$ pfdlint -f <project>/project_pessimistic.json
```

Report starting from the error / warning counts, listing each issue's severity, target ID, message, and a fix candidate. Do not delete `answers.md` while errors remain (so as not to lose the source data while the migration is incomplete).

## Finish

Once lint passes, delete `answers.md`.

```console
$ rm <project>/answers.md
```

Report to the user what was migrated (which keys went where, the settings filled in and their values, and the files deleted), and tell them that the scenario settings files are the reference from now on. Then hand off to `pjm-pfd-estimation-plan`.

If the master `ap.tsv` still has `楽観作業量` / `悲観作業量` (optimistic / pessimistic work volume) columns (and the PFD has no estimate boxes), the estimation method is also in the old format. That is a separate migration from `answers.md`, so tell the user to refer to the compatibility procedure in Derive Scenario Tables of `pjm-pfd-estimation-plan`.
