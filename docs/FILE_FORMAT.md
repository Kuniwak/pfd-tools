Deliverable File Definitions and Creation Steps
===============================================

The input data for pfd-tools consists of the PFD body, the various TSV tables that support the PFD, and a configuration JSON that ties them together.

## File list

| File | File format | Role |
| --- | --- | --- |
| PFD | draw.io XML file (`.drawio`) or a PNG exported from draw.io (`.drawio.png`) | Defines the dependencies between processes and deliverables |
| Atomic process table | TSV format | Defines the work volume, needed resources, and start condition for each process |
| Deliverable table | TSV format | Defines the available time and feedback limits for deliverables |
| Resource table | TSV format | Defines the dictionary of resource IDs |
| Composite deliverable table | TSV format | Defines deliverable sets that bundle multiple deliverables |
| Composite process table | TSV format | Defines higher-level processes that bundle multiple processes |
| Milestone table | TSV format | Defines the ordering relationships between milestones |
| Group table | TSV format | Defines the dictionary of groups |
| Classification table | TSV format | Assigns atomic processes to rows and bars of the master schedule (input to `planmaster`; an atomic process table with `MasterRow`/`MasterBar` columns can be used as is) |
| Project configuration file | JSON format | Bundles the individual definition files and passes them to the tool |
| Critical path analysis result | TSV format | Stores the output of the `criticalpath` command |

All TSV tables share two rules. First, the fixed columns of each table (`ID` and `Description`; additionally `Deliverables` for the composite deliverable table, and `Groups` and `Successors` for the milestone table) cannot be omitted. A table missing a fixed column is an error (reading it with missing columns would shift the positions of the ID and description). Second, you cannot write two or more rows with the same `ID`. Duplicate rows are an error (because there is no way to decide which row's values to use). This situation easily arises when two tables are simply concatenated, so edit tables by appending rows rather than concatenating them.

The minimal configuration consists of the PFD, the atomic process table, and the deliverable table (the finite resource model `-res finite` also needs the resource table). With just these you can run `pfdlint` and `pfdplan`. The composite deliverable table and the composite process table are optional; if they are not given, empty tables are assumed (the PFD is treated as having no composite deliverables or composite processes). If you use the master schedule (`planmaster`), you additionally need a classification table (if you add `MasterRow`/`MasterBar` columns to the atomic process table, you can pass the atomic process table as the classification table as is). The milestone table and the group table can optionally be passed to `planmaster` as display descriptions (`-row-meta` / `-bar-meta`).

## 1. Project configuration file

An execution configuration that bundles the paths to the individual definition files. Specify it with the `-f` option of `pfdlint`, `pfdplan`, `criticalpath`, `planmaster`, and others.

```json
{
  "pfd": "pfd.drawio",
  "atomic_process_table": "ap.tsv",
  "atomic_deliverable_table": "ad.tsv",
  "resource_table": "r.tsv",
  "composite_process_table": "cp.tsv",
  "composite_deliverable_table": "cd.tsv",
  "milestone_table": "m.tsv",
  "group_table": "g.tsv",
  "resource_mode": "finite",
  "feedback_mode": "enabled",
  "output_format": "mermaid",
  "start_day": "2026-05-20",
  "duration": 8,
  "not_biz_days": "holidays_ja.txt"
}
```

| Key | Type | Meaning | Required |
| --- | --- | --- | --- |
| `pfd` | string  | Path to the PFD body | ✅ |
| `atomic_process_table` | string  | Path to the atomic process table | ✅ |
| `atomic_deliverable_table` | string  | Path to the deliverable table | ✅ |
| `resource_table` | string  | Path to the resource table | ❌ (required when `resource_mode` is `finite`) |
| `composite_process_table` | string  | Path to the composite process table | ❌ |
| `composite_deliverable_table` | string  | Path to the composite deliverable table | ❌ |
| `milestone_table` | string  | Path to the milestone table | ❌ (read by `pfdlint`, `pfdtable`, and `pfdquery`) |
| `group_table` | string  | Path to the group table | ❌ (read by `pfdlint`, `pfdtable`, and `pfdquery`) |
| `resource_mode` | string  | How resources are treated (`finite` / `infinite`) | ❌ (default is `infinite`; see [Execution model](../README.md#execution-model)) |
| `feedback_mode` | string  | How feedback edges are treated (`enabled` / `disabled`) | ❌ (default is `disabled`; see [Execution model](../README.md#execution-model)) |
| `output_format` | string  | Output format of the execution plan (same values as `-out-format`) | ❌ (read by `pfdplan`, `plantimeline`, and `planmaster`) |
| `emphasis_tsv_path` | string  | Path to the emphasis ID table (same as `-em-tsv`) | ❌ (same as above; see "[11. Emphasis ID table](#11-emphasis-id-table)") |
| `start_day` | string  | Start date of the execution plan (YYYY-MM-DD; same as `-start`) | ❌ (same as above) |
| `start_time` | string  | Start time of a business day (same as `-start-time`) | ❌ (same as above; default is `10:00`) |
| `duration` | number  | Number of hours in a business day (same as `-duration`) | ❌ (same as above; default is 9) |
| `weekdays` | string  | Business days of the week (same as `-weekdays`) | ❌ (same as above; default is `mon,tue,wed,thu,fri`) |
| `not_biz_days` | string  | Path to the holiday list file (same as `-not-biz-days`) | ❌ (same as above) |
| `plan` | string  | Path to the execution plan (`plan-json`) | ❌ (read by `planmaster` and `plantimeline`; same as `<plan>` on the command line) |

The `planmaster` buffer factor (`-b`) is often varied across scenarios, so there is no configuration file key for it (specify it on the command line).

`output_format` and the business-hours settings (`start_day`, `start_time`, `duration`, `weekdays`, `not_biz_days`) are read by `pfdplan`, `plantimeline`, and `planmaster`. Values explicitly specified on the command line take precedence over the configuration file. Combining business-hours settings with an output format that does not use business hours (`plan-json`, `timeline-json`) is an error. Also, `planmaster` does not accept `plan-json` or `timeline-json`, since they have no corresponding master schedule output.

The same rule applies to the emphasis ID table (`emphasis_tsv_path`). Explicitly specifying `-em-tsv` on the command line together with `plan-json` or `timeline-json` is an error, but a value from the configuration file is simply unused for these output formats (this enables a two-stage workflow: produce `plan-json` with one configuration file, then render that `plan-json` with emphasis using `plantimeline`).

When handling multiple scenarios, such as optimistic/pessimistic or three-point estimates (optimistic, most likely, pessimistic), prepare one configuration file per scenario (e.g., `project_optimistic.json`, `project_pessimistic.json`). Usually only `atomic_process_table` and `resource_table` differ between scenarios; add files when adding scenarios.

Values in the configuration file can be overridden by command-line options (paths on the command line are resolved relative to the current directory at runtime; paths inside the configuration file are resolved relative to the directory containing the configuration file).

pfd-tools ignores unknown keys, so you may add information external to the tools (such as estimation assumptions) to this configuration file.

## 2. PFD body

A draw.io XML file. It defines the dependencies between processes and deliverables as a diagram.

All commands that take a PFD as input (`pfdrenum`, `pfdfix`, `pfdcpcomp`, `pfdestcallout`, `pfdesttable`, `pfdlint`, `pfdplan`, `pfdtable`, `pfdquery`, `pfddiff`, `plantimeline`, `criticalpath`, etc.) accept, in addition to `.drawio` (XML), a PNG exported from draw.io (`.drawio.png`). For a PNG, the mxfile embedded in the file is extracted and processed. `pfdrenum`, `pfdfix`, `pfdcpcomp`, and `pfdestcallout` return PNG output for PNG input (only the mxfile tEXt chunk is rewritten; the image itself is left unchanged).

| Element | Representation in draw.io | Meaning |
| --- | --- | --- |
| Atomic process | Oval with a thin border or no border (strokeWidth<=1) | An indivisible unit of work |
| Composite process | Oval with a thick border (strokeWidth>1) | A higher-level process that bundles multiple processes |
| Atomic deliverable | Rectangle with a thin border or no border (strokeWidth<=1) | A deliverable that is an input or output of a process |
| Composite deliverable | Rectangle with a thick border (strokeWidth>1) | A deliverable set that bundles multiple deliverables |
| Edge | Solid arrow | Input to or output from a process |
| Feedback edge | Dashed arrow | A dependency representing rework. It always goes from a deliverable to a process reachable in the reverse direction |

Node labels are managed in the form `P<number>: description` or `D<number>: description`. `P` IDs correspond to the atomic process table, and `D` IDs correspond to the deliverable table and the composite deliverable table. IDs with `.number` sub-numbers appended, such as `D3.1`, are also valid (`pfdrenum` does not generate them, but preserves them if assigned by hand).

The separator is the first half-width colon, so if an unnumbered title contains a half-width colon, the part before the colon is interpreted as the ID and is excluded from numbering by `pfdrenum`. In that case, `pfdlint` warns with `malformed-id`. If you want a title containing a colon, use the full-width colon `：`.

One ID is one element. Drawing the same ID as different kinds of elements is an error (a rectangle/oval mismatch, an atomic/composite deliverable mismatch, or an atomic/composite process mismatch). This applies to numbered IDs; unnumbered labels are distinguished by their appearance. An oval and a rectangle with the same label are numbered as separate elements, so when a gerund and a noun have the same form, you can write `(implement hoge) → [implement hoge]` (a thin/thick border mismatch is an error even when unnumbered). Note that a thick-bordered oval is treated as an atomic process if there is no detail page with the same name, so in that case mixing it with thin-bordered ovals is not an error. When drawing the same deliverable on both the parent diagram and a detail page, such as a boundary deliverable of a composite process, make the ID, description, and appearance all identical.

You may draw the same deliverable as multiple cells on one page (duplicate display). This is a way of cutting long edges short to make the diagram easier to read, and `pfdsort` creates them automatically. In this case, write the duplicate cells with ＊ after the ID, as in `D10＊: description`, and leave only the original (the cell that has the edge from the process that outputs the deliverable) unmarked. `pfddupmark` adds and removes ＊, so you do not need to manage it by hand. When reading, the half-width `*` is also accepted, and IDs with ＊ are read as `D10` by all tools. A page with a duplicate display always has the original on the same page.

Disable draw.io's compressed save (File > Properties > Compressed). In a compressed file, the page body becomes a base64 blob, and pfd-tools cannot read the diagram (unreadable pages are reported as errors). Also, every page needs an id attribute and a name attribute.

The work volume, resources, and start conditions needed for schedule computation are not held on the PFD; they are managed in the TSV tables.

## 3. Atomic process table

A TSV that defines the estimates and execution conditions of each process. The fixed columns are the two columns `ID` and `Description`; the columns after them are project-specific headers (ExtraHeaders). The tools detect columns by matching header names, so the column order is free. Japanese headers are also supported.

Typical column layout:

```tsv
ID	Description	Est. Work Volume	Est. Rework Volume Ratio	Needed Resources	Start Condition	MasterRow	MasterBar
```

| Column name (English / Japanese) | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string  | Process ID (corresponds to `Pxx` in the PFD) | ✅ | Fixed column |
| `Description` | string  | Process name | ✅ | Fixed column |
| `Est. Work Volume` / `予想作業量` | float64 (≧ 0)  | Estimated work volume (abstract unit) | ✅ | `pfdplan` computes the duration by dividing by the resource consumption rate |
| `Est. Rework Volume Ratio` / `予想手戻り作業量割合` | float64 (0 ≦ ratio ≦ 1)  | Rework ratio (0 to 1) as a floating-point number | ✅ when `-fb enabled` | Rework volume decays exponentially as `initial work volume × ratio^(number of reworks)`. See [Execution model](../README.md#execution-model) |
| `Needed Resources` / `必要資源` | string  | Allocation of needed resources | ✅ when `-res finite` | See the format below. See [Execution model](../README.md#execution-model) |
| `Start Condition` / `開始条件` | string  | Start condition | ❌ | Empty means unconditional. See the syntax below |
| `MasterBar` | string  | Bar key of the master schedule (e.g., a milestone ID) | ❌ | Same column name as the classification table. Locale-independent |
| `MasterRow` | string  | Row key of the master schedule (e.g., a group ID) | ❌ | Same column name as the classification table. Multiple values may be comma-separated. Locale-independent |

### Needed Resources format

Multiple resource allocations are separated by semicolons `;`. Each allocation has the form `ResourceID:work volume consumed per unit time`. The work volume consumed per unit time may be a decimal.

```
WebApp:1            # Allocate WebApp with work volume consumed per unit time of 1
WebApp:1;QA:0.5     # Allocate WebApp and QA simultaneously (QA with work volume consumed per unit time of 0.5)
```

Listing resource IDs separated by commas `,` means that all of these resources are occupied.

```
WebApp,iOSApp:1      # Allocate both WebApp and iOSApp with work volume consumed per unit time of 1
```

### Start Condition syntax

A process becomes executable if and only if all of the following are true:

* All input deliverables of the process have been generated (feedback deliverables may or may not exist)
* Among the input deliverables of the process, including feedback deliverables, there is one that has been updated since it was last processed
* The Start Condition expression of the process evaluates to true

```
precondition = *SP *1(or_expr)
or_expr      = and_expr *( "||" *SP and_expr)
and_expr     = primary  *( "&&" *SP primary )
primary      = "(" *SP precondition ")" *SP
             / "\execBetween(" *SP node_id *SP "," *SP bound_expr *SP "," *SP bound_expr *SP ")" *SP
             / "\complete(" *SP ("*" / node_id) *SP ")" *SP
             / "\exec(" *SP node_id *SP ")" *SP
             / "!" *SP precondition *SP
bound_expr   = int / "\inf" / "\maxRev(" *SP node_id *SP ")" *SP
node_id      = *(DIGIT / ALPHA / "_" / "-" / ".") 1*(DIGIT / ALPHA)
```

| Expression | Meaning |
| --- | --- |
| (empty) | True (`\true`) |
| `\execBetween(D2, 0, 3)` | True only while the revision of deliverable D2 is in the half-open interval `[0, 3)` |
| `\execBetween(D2, \maxRev(D2), \inf)` | True only while the revision of deliverable D2 is at least `\maxRev(D2)` |
| `\complete(D2)` | Syntactic sugar for `\execBetween(D2, \maxRev(D2), \inf)` |
| `\complete(*)` | False until all feedback source deliverables reachable in the reverse direction are complete; true afterwards |
| `\exec(P3)` | False until process P3 becomes executable; true once it becomes executable |
| `A && B` | True if both A and B are true; false otherwise |
| `A \|\| B` | True if A or B is true; false otherwise |
| `!A` | True if A is false; false otherwise |

A `revision` of `0` means not yet generated, and `1` means generated for the first time. `\inf` used in `Start Condition` is a reserved word, and `\maxRev(Dx)` refers to `Max Revision` in the deliverable table.

## 4. Deliverable table

A TSV that defines additional attributes of the deliverable nodes (`Dxx`) in the PFD. The fixed columns are `ID` and `Description`; the rest are ExtraHeaders.

Typical column layout:

```tsv
ID	Description	Available Time	Max Revision
```

| Column name (English / Japanese) | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string  | Deliverable ID (corresponds to `Dxx` in the PFD) | ✅ | Fixed column |
| `Description` | string  | Deliverable name | ✅ | Fixed column |
| `Available Time` / `利用可能時刻` | float64  | Time at which the initial deliverable becomes available | ❌ | Effective only for deliverables with no upstream process. Empty/0 means "available from the start" |
| `Max Revision` / `最大版` | int  | Termination condition of the feedback loop | ✅ when `-fb enabled` | Set only for feedback source deliverables (integer ≥ 1). The loop ends when revision ≥ Max Revision. Leave empty or `-` for non-feedback deliverables. Not read with `-fb disabled` |

## 5. Resource table

A dictionary of resource IDs referenced by `Needed Resources` in the atomic process table. Used only with `-res finite` (the finite resource model).

```tsv
ID	Description
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Resource ID (referenced by `Needed Resources`) | ✅ | Fixed column |
| `Description` | string | Description of the resource (e.g., a person or team name) | ✅ | Fixed column |

## 6. Composite deliverable table

A TSV that defines deliverable sets that bundle multiple deliverables. This table is optional; if not given, an empty table (no composite deliverables) is assumed. However, if the PFD diagram has composite deliverable nodes, a row corresponding to each composite deliverable is required. If there is no corresponding row, an error stating that the composite deliverable table has no entry (`missing-cd-table`) is reported.

```tsv
ID	Description	Deliverables
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Composite deliverable ID | ✅ | Fixed column |
| `Description` | string | Description of the composite deliverable | ✅ | Fixed column |
| `Deliverables` | string | Comma-separated list of contained deliverable IDs (e.g., `D1, D2, D3`) | ✅ | Fixed column |

The breakdown may mix atomic deliverable IDs and composite deliverable IDs. Writing a composite deliverable ID lets you nest composite deliverables in multiple levels; the set of atomic deliverables that such a composite deliverable represents is the set of atomic deliverables obtained by recursively expanding the contained composite deliverables. There is no limit on nesting depth, but a cycle in which a composite deliverable transitively contains itself is an error (`pfdlint` reports it as `acyclic-cd-comp`, and other tools fail because they cannot interpret the PFD).

Nesting is expressed in this table, not in the diagram, so composite deliverables used only as parts of a breakdown need not be drawn in the diagram (IDs reached by following breakdowns from a composite deliverable drawn in the diagram are not reported as `extra-cd-table`). Conversely, rows not reachable from any composite deliverable drawn in the diagram are stale unused rows, so they are reported together as `extra-cd-table` even if they are chained. Rows remaining for IDs drawn in the diagram as atomic deliverables (thin-bordered rectangles) are also reported as stale.

Draw every atomic deliverable listed in a breakdown as a node somewhere in the diagram. An edge connected to a composite deliverable is expanded into edges connected to each deliverable in the breakdown, so if one is not drawn, an `in-field` error occurs.

## 7. Composite process table

A TSV that defines higher-level processes that bundle multiple processes.

```tsv
ID	Description
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Composite process ID | ✅ | Fixed column |
| `Description` | string | Description of the composite process | ✅ | Fixed column |

Which atomic processes a composite process contains is expressed not in this table but in a draw.io detail page (page name = composite process ID). The processes drawn on the detail page become members of that composite process.

Page names can be written in the same `ID: description` notation as process labels. So that you can tell which process a page breaks down just by its name, a detail page can be named `P123: foo bar` and the context diagram `P0: Context diagram` (an ID alone, such as `P123` or `P0`, still works as before). The separator is the first half-width colon, and surrounding whitespace is ignored. Since the meaning of the diagram is determined by the ID part alone, the description part is for human readers. If the description disagrees with the process label, a warning is issued because it may be a missed update (it is not an error).

By drawing another composite process (a thick-bordered oval with its own detail page) on a detail page, you can nest composite processes in multiple levels. The atomic processes contained in a nested composite process are the set of atomic processes obtained by recursively expanding the lower-level composite processes. There is no limit on nesting depth, but a cycle in which a composite process transitively contains itself is an error.

## 8. Milestone table

A TSV that defines the ordering relationships between milestones. Besides being used for consistency checks by `pfdlint`, passing it to `planmaster`'s `-bar-meta` makes it the display description of bars (`planmaster` does not read the `Groups` and `Successors` columns).

```tsv
ID	Description	Groups	Successors
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Milestone ID | ✅ | Fixed column |
| `Description` | string | Description of the milestone | ✅ | Fixed column |
| `Groups` | string | Comma-separated list of group IDs belonging to this milestone | ✅ | Fixed column |
| `Successors` | string | Comma-separated list of successor milestone IDs | ✅ | Fixed column |

## 9. Group table

A dictionary of group IDs. Referenced by the `MasterRow` column of the atomic process table and the `Groups` column of the milestone table. Passing it to `planmaster`'s `-row-meta` makes it the display description of rows.

```tsv
ID	Description
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Group ID | ✅ | Fixed column |
| `Description` | string | Description of the group | ✅ | Fixed column |

## 10. Critical path analysis result

The output of the `criticalpath` command. It is not an input definition file.

The headers switch according to `-locale` (default is `ja`).

```tsv
ID	最大弾性値（全余裕）	最小弾性値
```

Specifying `-locale en` gives English headers.

```tsv
ID	Maximum Elasticity (Total Float)	Minimum Elasticity
```

| Column name (ja / en) | Type | Meaning |
| --- | --- | --- |
| `ID` | string | Atomic process ID. Since it is the same ID column as in the other tables, a filtered version of this output can be passed directly as the emphasis ID table (`-em-tsv`) |
| `最大弾性値（全余裕）` / `Maximum Elasticity (Total Float)` | float64 (≧ 0) | Maximum elasticity. The maximum amount by which the duration of the process can be extended without changing the overall duration; corresponds to the Total Float in CPM (processes with `0.00` are candidates on the critical path) |
| `最小弾性値` / `Minimum Elasticity` | float64 (> 0) or `"-"` | Minimum elasticity. The amount by which the overall duration is shortened when the duration of the process is reduced to its minimum; the larger the value, the more room for shortening the whole (`-` means the whole does not change even if shortened) |

## 11. Emphasis ID table

A TSV passed to `-em-tsv` of `plantimeline` / `pfdplan` / `planmaster`. Only the bars with the IDs listed here are emphasized (the `crit` tag in mermaid, color in PlantUML, and an extra trailing `Emphasis` column in `google-spreadsheet-tsv`).

```tsv
ID
P1
P3
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | ID to emphasize | ✅ | Atomic process ID for `plantimeline` / `pfdplan`; bar ID (value of the `MasterBar` column of the classification table) for `planmaster` |

The `ID` column is identified by name, so it can be in any position, and other columns are ignored. Therefore, you can pass the result of filtering "10. Critical path analysis result" with qhs as is (the examples below use the default `ja` header `最大弾性値（全余裕）`, i.e. Maximum Elasticity (Total Float)).

```console
$ criticalpath -poor -f path/to/project.json >cp.tsv
$ qhs -H -O -t -T 'SELECT ID FROM cp.tsv WHERE "最大弾性値（全余裕）" < 0.0001' >em.tsv
$ plantimeline -f path/to/project.json -out-format mermaid -em-tsv em.tsv path/to/plan.json
```

Deciding what to emphasize (being on the critical path, being delayed, etc.) is outside the tools. By creating a table with an `ID` column containing only the rows you want to emphasize, you can emphasize by any criterion.

IDs not present in the target are ignored with a warning (can be suppressed with `-silent`). Since emphasis is a visual specification, specifying `-em-tsv` together with `-out-format plan-json` / `timeline-json`, which represent the execution plan itself, is an error (`emphasis_tsv_path` in the configuration file is simply unused).

Since the emphasis target of `planmaster` is the bar ID, if you want to emphasize the bars that contain atomic processes on the critical path, join with the `MasterBar` column of the classification table.

```console
$ qhs -H -O -t -T 'SELECT DISTINCT t1.MasterBar AS ID FROM path/to/master.tsv AS t1 JOIN cp.tsv AS t2 ON t1.ID = t2.ID WHERE t2."最大弾性値（全余裕）" < 0.0001 AND LENGTH(t1.MasterBar) > 0' >em.tsv
```

## 12. Classification table

A TSV passed to `-tsv` of `planmaster`. It assigns atomic processes to rows (`MasterRow`) and bars (`MasterBar`) of the master schedule. Columns are identified by name and unknown columns are ignored, so you can pass an atomic process table that has `MasterRow`/`MasterBar` columns as is. The axis for grouping is up to you; the table can be written by hand or produced by processing with `qhs` / tsv-tools.

```tsv
ID	MasterRow	MasterBar
P1	G1	M1
P2	G1,G2	M2
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Atomic process ID | ✅ | An ID not in the execution plan is an error. Two or more rows with the same ID are an error |
| `MasterRow` | string | Comma-separated list of row keys | ✅ | One atomic process can be placed in multiple rows |
| `MasterBar` | string | Bar key | ✅ | Exactly one |

Columns are identified by name, so their positions are free, and other columns are ignored. Not every atomic process needs to be classified (unclassified ones are not drawn, so you can draw only a part by narrowing the table). Rows where any of `ID`, `MasterRow`, or `MasterBar` is empty are also treated as "unclassified" and not drawn. For the drawing rules (first execution interval only, folding into min/max), see `planmaster -h`.

## Creation steps

The steps to create a full PFD set for a new project.

### Step 1: Draw the PFD

Create the PFD in draw.io.

- Draw processes as ovals and deliverables as rectangles
- Draw atomic elements with a thin border or no border (`strokeWidth<=1`) and composite elements with a thick border (`strokeWidth>1`)
- Make edges representing rework dashed
- Write an ID and a description in the label, such as `P1: Design`, `D1: Design document`

If there are nodes whose IDs are not yet assigned, you can number them automatically with `pfdrenum`. A process and a deliverable with the same label are numbered as separate elements, receiving `P%d` and `D%d` respectively:

```console
$ pfdrenum -inplace path/to/pfd.drawio
```

If you split the PFD into multiple files and combine them with `drawiocat`, separate deciding the numbering (for the whole) from applying it (per file). For details, see the pfdrenum section of the README:

```console
$ drawiocat *.drawio | pfdrenum -out-format tsv >renum.tsv
$ for f in *.drawio; do pfdrenum -renum-plan renum.tsv -inplace "$f"; done
$ drawiocat *.drawio | pfdrenum -out-format tsv   # 0 rows means the numbering has been applied
```

If there are variations that replace the breakdown (detail page) of a composite process, pass the variations through the numbering plan one at a time so that numbering is unique across all variations. For details, see "Numbering across variations" in the pfdrenum section of the README.

### Step 2: Generate templates for the TSV tables

Generate templates for each table from the PFD with the `pfdtable` command:

```console
$ pfdtable -t cd -p path/to/pfd.drawio > cd.tsv
$ pfdtable -t ap -p path/to/pfd.drawio > ap.tsv
$ pfdtable -t ad -p path/to/pfd.drawio > ad.tsv
$ pfdtable -t r  -p path/to/pfd.drawio > r.tsv   # needed only for the finite resource model (-res finite)
```

With `-t all`, you can generate all tables and the configuration JSON at once. If you pass the execution model, columns and tables that the model does not read are not generated:

```console
$ pfdtable -t all-plan -res finite -fb enabled -p path/to/pfd.drawio -out-dir path/to/project
```

When the PFD is updated, the `-existing` option lets you update while preserving the existing table:

```console
$ pfdtable -t ap -existing path/to/ap.tsv -p path/to/pfd.drawio > ap_new.tsv
```

### Step 3: Fill in the TSV tables

Add the necessary columns to the generated templates and fill in the values.

- **Atomic process table** (`ap.tsv`): Add an `Est. Work Volume` column and fill in the estimate for each process. For the finite resource model, also add `Needed Resources`; for the model with rework, also add `Est. Rework Volume Ratio` (see [Execution model](../README.md#execution-model)). Add `Start Condition`, `Milestone`, and `Group` columns as needed
- **Deliverable table** (`ad.tsv`): Add an `Available Time` column and fill in the attributes of the necessary deliverables. For the model with rework, also add `Max Revision`
- **Resource table** (`r.tsv`): Finite resource model only. Fill in the IDs and descriptions of the resources (teams, people) used in the project

### Step 4: Create the configuration JSON

Create a configuration JSON that bundles the paths to each file:

```json
{
  "pfd": "pfd.drawio",
  "atomic_process_table": "ap.tsv",
  "atomic_deliverable_table": "ad.tsv",
  "resource_table": "r.tsv",
  "composite_deliverable_table": "cd.tsv",
  "resource_mode": "finite",
  "feedback_mode": "enabled"
}
```

If `resource_mode` / `feedback_mode` are omitted, the default execution model (infinite resources, no rework) is used, and `resource_table` is also unnecessary.

### Step 5: Run the static check

Check the consistency of the PFD and each table with `pfdlint`:

```console
$ pfdlint -f ./project.json
```

Fix the PFD and TSVs until there are no errors.

### Step 6: Generate the schedule

Generate the execution plan with `pfdplan`:

```console
$ pfdplan -f ./project.json -poor -start 2025-06-01 -start-time 10:00 -duration 9 \
    -not-biz-days <(holidays -locale ja)
```

`-poor` (greedy method) chooses the transition that consumes the most work volume at each step, but when multiple transitions tie, it picks one using a pseudo-random number based on `-random-seed`. If `-random-seed` is omitted, different random numbers are used on each run, so results may vary. To reproduce the same plan, specify a fixed value for `-random-seed`.

If the project is so large that even `-poor` does not finish, use `-poorest`. `-poorest` does not enumerate all allocations at each step; it greedily constructs exactly one directly, so it finishes in a practical time even with many allocatable processes and resources. Its behavior is deterministic, so `-random-seed` does not affect the result. However, since it does not necessarily choose the optimal allocation at each step, plan quality may be worse than with `-poor`.

### Step 7: Analyze the critical path

Identify bottlenecks with `criticalpath` (besides `-poor` there are also `-best` and `-better`, but they usually do not finish, so `-poor` is recommended):

```console
$ criticalpath -f ./project.json -poor
```
