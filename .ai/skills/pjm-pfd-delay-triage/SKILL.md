---
name: pjm-pfd-delay-triage
description: |
  A lightweight triage for when a worker (process owner) thinks "my process is going to be late". It uses the total float (maximum elasticity) from the critical path to judge the impact on the planned end date and decides whether escalation to the PjM/PMO is needed. If the delay is within the total float, no escalation is needed and the delay is OK; if it exceeds it, it generates an escalation report addressed to the PjM/PMO. Triggers when the user says things like "it's going to be late", "I might not make it", "how much can I delay", "I want to check whether escalation is needed" (or in Japanese, "遅れそう", "間に合わなそう", "どこまで遅らせていい", "エスカレーションが必要か確認したい"). It does not replan (redo the dates) itself; the PjM/PMO does that with pjm-pfd-replan.
allowed-tools: Read, Write, Bash, Glob, Grep, readfile, writefile, terminal, glob, grep
---

# pjm-pfd-delay-triage

In a project planned with `pjm-pfd-estimation-plan`, when a worker expects a delay in a process they own, judge from the critical path's total float whether that delay moves the project's planned end date. If it is within the total float, the delay is OK without escalation; if it exceeds it, generate an escalation report to the PjM/PMO. This is the triage part of `pjm-pfd-replan`, split out so a worker can run it alone; this skill does not replan (re-derive the scenario tables or redo the dates).

## Workflow

- [ ] 1. Confirm the working language: ask the user whether to proceed in 日本語 or English. Use the chosen language for all subsequent conversation, explanations, and progress reports. (Command names, paths, option flags, column names, and code stay verbatim regardless of the chosen language.)
- [ ] 2. Check the prerequisites ([Prerequisites](#prerequisites))
- [ ] 3. Collect and validate the delayed process and the number of days of delay ([Input](#input))
- [ ] 4. Compare the delay with the total float ([Float Check](#float-check))
- [ ] 5. Check the impact scope ([Impact Scope](#impact-scope))
- [ ] 6. Check the risk to external commitments ([Commitment Check](#commitment-check))
- [ ] 7. Report the verdict and, if needed, generate an escalation report ([Verdict And Report](#verdict-and-report))

## Conventions

- At the start, confirm the working language (日本語 / English) with the user, then use it consistently for all conversation, explanations, and progress reports. Keep command names, paths, options, and column names verbatim.
- Judge using the **scenario the plan's dates are based on** (default: optimistic). The command examples in this document are written for optimistic; if the base scenario is pessimistic, read `optimistic` in file and setting names as `pessimistic`.
- The search-quality fallback (if `-poor` does not finish within 20 seconds, use `-poorest`; once you switch for one, switch for all subsequent runs) follows the Conventions of `pjm-pfd-estimation-plan`.
- This skill only judges; it does not change plan artifacts (scenario tables, plan-json, timelines). The only things it may generate are `criticalpath_*.tsv` (if missing) and the escalation report.
- Depending on the user's request, you may need to investigate changes, successor processes, and so on; pfd-tools can help. In that case, run `pfdhelp -short` to see a description of the pfd-tools tools.

## Prerequisites

Check the following. If anything is missing, run `pjm-pfd-estimation-plan` first (or `pjm-pfd-replan` if progress has already been reflected):

- A planned project: `<project>/project_optimistic.json` / `project_pessimistic.json` and the scenario tables they point to (`ap_optimistic.tsv` / `ap_pessimistic.tsv`, etc.)
- `criticalpath_optimistic.tsv` for the base scenario (if missing, generate it at the start of [Float Check](#float-check))
- The scenario tables belong to the current plan (they have not been changed since the last estimation-plan / replan). If they have been changed, the plan underlying the dates and the analysis target are out of sync, so the prerequisite is not met; run `pjm-pfd-replan` first

## Input

Ask the worker for:

- The ID of the process likely to be delayed (e.g. `P12`)
- The expected delay in business days

If they give a "new expected end date" instead, convert it into the number of business days from that process's planned end date in the old plan to the new expected end date. Read the planned end date from `timeline_optimistic.tsv` (aggregated per process: start = first start, end = last end; for the qhs aggregation command, see Timeline Diff in `pjm-pfd-replan`) or from the Gantt, and exclude holidays using the holiday list used at planning time (the file pointed to by `not_biz_days` in the scenario settings file; default `<project>/holidays_ja.txt`). If you do not use the same list as the plan, the business-day count will be off.

Validation and branching:

- Confirm the ID exists in the `ID` column of the base scenario's scenario table (`ap_optimistic.tsv`) (if not, confirm the input again)
- **If multiple processes are delayed at the same time, proceed at this point to the escalation in [Verdict And Report](#verdict-and-report).** Elasticity is shared among processes on the same path, so you must not judge by adding or subtracting it. Include each process's total float and impact scope in the report as reference information

## Float Check

Use the base scenario's `criticalpath_*.tsv`. If missing, generate it now (it will be an analysis of the current plan; if `-poor` does not finish within 20 seconds, see [Notes](#notes)):

```console
$ criticalpath -f <project>/project_optimistic.json -poor > <project>/criticalpath_optimistic.tsv
```

Compare `最大弾性値（全余裕）` (maximum elasticity / total float) in the delayed process's row with the delay in business days:

- Delay in business days ≤ total float → the base scenario's planned end date (the basis of ticket due dates) does not move
- Delay in business days > total float → the planned end date is expected to move

The total float is in the same business-day unit as the estimate (`予想作業量`). Because it is relative to a greedy-search plan it can be negative; treat negative values as 0 (no total float).

## Impact Scope

Check the successor processes whose dates may move:

```console
$ pfdquery -f <project>/project_optimistic.json -reachable <delayed process ID>
```

The `REACHABLE` rows list the delayed process itself and the successor processes reachable from it. **Pass only a process ID to `-reachable`** (there is a known issue where passing a deliverable ID causes a panic).

## Commitment Check

For the risk to external commitments (dates promised on a pessimistic basis), make the same judgment with the total float in `criticalpath_pessimistic.tsv` (if missing, generate it as in [Float Check](#float-check)). If exceeded, include in the report that a date adjustment with stakeholders is likely to be needed. If the base scenario is pessimistic, this check is identical to [Float Check](#float-check) and is skipped.

## Verdict And Report

| Situation | Verdict |
|---|---|
| Single-process delay, delay in business days ≤ total float (base scenario) | No escalation needed (delay OK) |
| Delay in business days > total float | Escalate |
| Multiple processes delayed at the same time | Escalate |
| The estimate itself was revised (assumptions changed) | Escalate |

Even when no escalation is needed, report the following to the worker:

- The basis of the verdict (delay in business days, total float, base scenario) and the result of [Commitment Check](#commitment-check)
- Caution: even within the total float, the start and end dates of serial successors (the list from [Impact Scope](#impact-scope)) move. Share the delay with the owners of the successors. The date shifts are updated at the next progress reflection / replan (`pjm-pfd-replan`)

When escalating, generate a report addressed to the PjM/PMO and present it to the worker. This skill does not send it (via chat, ticket comments, etc.); leave that to the worker:

```markdown
## Delay escalation: <process ID> (<description>)

- Project: <project>
- Expected delay: <N> business days (situation: <summary of what the worker said>)
- Verdict: total float <total float> business days < delay <N> business days, so the planned end date (<base scenario>) is expected to move
- External commitment (pessimistic): total float <total float> business days → <exceeded (a date adjustment with stakeholders is likely needed) / not exceeded>
- Impact scope (processes whose dates may move): <list from REACHABLE>
- Request: replan with `pjm-pfd-replan`
- Reference (suggested progress.tsv entry): <include only if the delayed process's remaining optimistic / pessimistic work volume is known>
```

When escalating because of simultaneous delays in multiple processes or a revised estimate, adapt the heading and the verdict line (list the expected delay, total float, and impact scope per process, and give the reason as "because multiple processes are delayed at the same time" or "because the estimation assumptions changed").

## Notes

- Elasticity values are relative to the previous plan and previous estimate, and are shared among processes on the same path. Never judge multiple delays by adding or subtracting elasticity (when in doubt, escalate and have it replanned).
- Triage is a judgment of end-date risk, not a way to skip date updates. Dates are recalculated at progress reflection / replan (`pjm-pfd-replan`). Do not manually slide all successors uniformly, as that ignores resource conflicts.
