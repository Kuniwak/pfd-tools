pfd-tools
=========

[![DeepWiki](https://img.shields.io/badge/DeepWiki-Kuniwak%2Fpfd--tools-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACwAAAAyCAYAAAAnWDnqAAAAAXNSR0IArs4c6QAAA05JREFUaEPtmUtyEzEQhtWTQyQLHNak2AB7ZnyXZMEjXMGeK/AIi+QuHrMnbChYY7MIh8g01fJoopFb0uhhEqqcbWTp06/uv1saEDv4O3n3dV60RfP947Mm9/SQc0ICFQgzfc4CYZoTPAswgSJCCUJUnAAoRHOAUOcATwbmVLWdGoH//PB8mnKqScAhsD0kYP3j/Yt5LPQe2KvcXmGvRHcDnpxfL2zOYJ1mFwrryWTz0advv1Ut4CJgf5uhDuDj5eUcAUoahrdY/56ebRWeraTjMt/00Sh3UDtjgHtQNHwcRGOC98BJEAEymycmYcWwOprTgcB6VZ5JK5TAJ+fXGLBm3FDAmn6oPPjR4rKCAoJCal2eAiQp2x0vxTPB3ALO2CRkwmDy5WohzBDwSEFKRwPbknEggCPB/imwrycgxX2NzoMCHhPkDwqYMr9tRcP5qNrMZHkVnOjRMWwLCcr8ohBVb1OMjxLwGCvjTikrsBOiA6fNyCrm8V1rP93iVPpwaE+gO0SsWmPiXB+jikdf6SizrT5qKasx5j8ABbHpFTx+vFXp9EnYQmLx02h1QTTrl6eDqxLnGjporxl3NL3agEvXdT0WmEost648sQOYAeJS9Q7bfUVoMGnjo4AZdUMQku50McDcMWcBPvr0SzbTAFDfvJqwLzgxwATnCgnp4wDl6Aa+Ax283gghmj+vj7feE2KBBRMW3FzOpLOADl0Isb5587h/U4gGvkt5v60Z1VLG8BhYjbzRwyQZemwAd6cCR5/XFWLYZRIMpX39AR0tjaGGiGzLVyhse5C9RKC6ai42ppWPKiBagOvaYk8lO7DajerabOZP46Lby5wKjw1HCRx7p9sVMOWGzb/vA1hwiWc6jm3MvQDTogQkiqIhJV0nBQBTU+3okKCFDy9WwferkHjtxib7t3xIUQtHxnIwtx4mpg26/HfwVNVDb4oI9RHmx5WGelRVlrtiw43zboCLaxv46AZeB3IlTkwouebTr1y2NjSpHz68WNFjHvupy3q8TFn3Hos2IAk4Ju5dCo8B3wP7VPr/FGaKiG+T+v+TQqIrOqMTL1VdWV1DdmcbO8KXBz6esmYWYKPwDL5b5FA1a0hwapHiom0r/cKaoqr+27/XcrS5UwSMbQAAAABJRU5ErkJggg==)](https://deepwiki.com/Kuniwak/pfd-tools)

Tools related to Process Flow Diagram (PFD) such as static analyzers and schedulers.

**:bee: This repository is not accepting pull requests.**


Installation
------------
Download the latest binary from [Releases](https://github.com/Kuniwak/pfd-tools/releases) and place it in a directory that is in your PATH.


Skills (Claude Code / Cursor)
-----------------------------

This repository ships two AI agent skills that automate the workflow described in [Input Files and Creation Steps](#input-files-and-creation-steps):

- **pjm-pfd-project-bootstrap** — scaffolds a new pfd-tools project from a `pfd.drawio.png` template and runs the pre-estimation lint gate.
- **pjm-pfd-estimation-plan** — produces a two-point (optimistic/pessimistic) estimate, analyzes the critical path, and generates a Mermaid Gantt or a Google Sheets Timeline.

Each skill asks whether to work in English or Japanese before it starts.

### Claude Code

```console
$ claude
> /plugin marketplace add Kuniwak/pfd-tools
> /plugin install pfd-skills@pfd-tools
```

### Cursor

The skills live under `.cursor/skills/` (and `.agents/skills/`), so Cursor picks them up automatically when this repository is open. To use them in another project, copy the `pjm-pfd-*` directories from [`.ai/skills/`](.ai/skills) into that project's `.cursor/skills/`.

Input Files and Creation Steps
------------------------------

The input data for pfd-tools consists of the PFD body, the various TSV tables that support the PFD, and a configuration JSON that ties them together.

### File list

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
| Project configuration file | JSON format | Bundles the individual definition files and passes them to the tool |
| Critical path analysis result | TSV format | Stores the output of the `criticalpath` command |

The minimal configuration consists of the PFD, the atomic process table, the deliverable table, the composite deliverable table, the composite process table, and the resource table. With just these you can run `pfdlint` and `pfdplan`. If you use the master schedule (`planmaster`), you additionally need the milestone table and the group table.

### 1. Project configuration file

This is an execution setting that bundles the paths of the individual definition files. You specify it with the `-f` option of commands such as `pfdlint`, `pfdplan`, and `criticalpath`.

```json
{
  "pfd": "pfd.drawio",
  "atomic_process_table": "ap.tsv",
  "atomic_deliverable_table": "ad.tsv",
  "resource_table": "r.tsv",
  "composite_process_table": "cp.tsv",
  "composite_deliverable_table": "cd.tsv",
  "milestone_table": "m.tsv",
  "group_table": "g.tsv"
}
```

| Key | Type | Meaning | Required |
| --- | --- | --- | --- |
| `pfd` | string  | Path to the PFD body | ✅ |
| `atomic_process_table` | string  | Path to the atomic process table | ✅ |
| `atomic_deliverable_table` | string  | Path to the deliverable table | ✅ |
| `resource_table` | string  | Path to the resource table | ❌ |
| `composite_process_table` | string  | Path to the composite process table | ❌ |
| `composite_deliverable_table` | string  | Path to the composite deliverable table | ✅ |
| `milestone_table` | string  | Path to the milestone table | ❌ (required when using `planmaster`) |
| `group_table` | string  | Path to the group table | ❌ (required when using `planmaster`) |

When you handle multiple estimates such as optimistic/pessimistic, prepare multiple configuration files that swap out only the `atomic_process_table` (e.g., `project1.json`, `project2.json`).

### 2. PFD body

This is a draw.io XML file. It defines the dependencies between processes and deliverables as a diagram.

All commands that take a PFD as input (`pfdrenum`, `pfdlint`, `pfdplan`, `pfdtable`, `pfdquery`, `pfddiff`, `plantimeline`, `criticalpath`, etc.) accept, in addition to `.drawio` (XML), a PNG exported from draw.io (`.drawio.png`) as input. For a PNG, the embedded mxfile is extracted from the file and processed. For PNG input, `pfdrenum` returns PNG output (it rewrites only the tEXt chunk of the mxfile and leaves the image itself unchanged).

| Element | Representation in draw.io | Meaning |
| --- | --- | --- |
| Atomic process | Thin-line ellipse (strokeWidth=1) | An indivisible unit of work |
| Composite process | Thick-line ellipse (strokeWidth>1) | A higher-level process that bundles multiple processes |
| Atomic deliverable | Thin-line rectangle (strokeWidth=1) | A deliverable that is an input/output of a process |
| Composite deliverable | Thick-line rectangle (strokeWidth>1) | A deliverable set that bundles multiple deliverables |
| Edge | Solid arrow | An input/output to a process |
| Feedback edge | Dashed arrow | A dependency that represents rework. It must always point from a deliverable to a process reachable in the reverse direction |

Node labels are managed in the form `P<number>: description` or `D<number>: description`. `P`-type IDs correspond to the atomic process table, and `D`-type IDs correspond to the deliverable table and the composite deliverable table.

The work volume, resources, and start conditions needed for schedule calculation are not stored on the PFD itself; they are managed in TSV tables.

### 3. Atomic process table

This is a TSV that defines the estimate values and execution conditions for each process. The fixed columns are the two columns `ID` and `Description`; the columns after them are project-specific headers (ExtraHeaders). Because the tool detects columns by matching header names, the column order is free. Japanese headers are also supported.

A typical column layout:

```tsv
ID	Description	Est. Work Volume	Est. Rework Volume Ratio	Needed Resources	Start Condition	Milestone	Group
```

| Column name (English / Japanese) | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string  | Process ID (corresponds to `Pxx` in the PFD) | ✅ | Fixed column |
| `Description` | string  | Process name | ✅ | Fixed column |
| `Est. Work Volume` / `予想作業量` | float64 (≧ 0)  | Estimated work volume (abstract unit) | ✅ | `pfdplan` divides this by the resource consumption rate to compute the required duration |
| `Est. Rework Volume Ratio` / `予想手戻り作業量割合` | float64 (0 ≦ ratio ≦ 1)  | A floating-point rework ratio (0–1) | ✅ | Rework volume = `initial work volume × ratio^(number of reworks)`, decaying exponentially |
| `Needed Resources` / `必要資源` | string  | Allocation of needed resources | ✅ | See the format described below |
| `Start Condition` / `開始条件` | string  | Start condition | ❌ | An empty cell means unconditional. See the syntax described below |
| `Milestone` / `マイルストーン` | string  | The milestone it belongs to | ❌ | Used by `planmaster` |
| `Group` / `グループ` | string  | The group it belongs to | ❌ | Used by `planmaster` |

#### Needed Resources format

Separate multiple resource allocations with a semicolon `;`. Each allocation is in the form `resourceID:work volume consumed per unit time`. The work volume consumed per unit time can also be a decimal.

```
CCWeb:1            # Allocate CCWeb with a work volume consumed per unit time of 1
CCWeb:1;QA:0.5     # Allocate CCWeb and QA simultaneously (QA has a work volume consumed per unit time of 0.5)
```

Listing resource IDs separated by a comma `,` means that all of these resources are occupied.

```
CCWeb,CCiOS:1      # Allocate both CCWeb and CCiOS with a work volume consumed per unit time of 1
```

#### Start Condition syntax

A process becomes executable if and only if all of the following are true:

* All input deliverables of that process have been produced (feedback deliverables may or may not be present)
* Among the input deliverables of that process, including feedback deliverables, there is one that has not been updated since the last update of the others
* The evaluation result of that process's Start Condition expression is true

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
| `\complete(*)` | False until all feedback-source deliverables reachable in the reverse direction are complete; true after they are complete |
| `\exec(P3)` | False until process P3 becomes executable; true once it becomes executable |
| `A && B` | True if both A and B are true; otherwise false |
| `A \|\| B` | True if A or B is true; otherwise false |
| `!A` | True if A is false; otherwise false |

A `revision` of `0` means not yet produced, and `1` means the first production. The `\inf` used in `Start Condition` is a reserved word, and `\maxRev(Dx)` references `Max Revision` in the deliverable table.

### 4. Deliverable table

This is a TSV that defines the additional attributes of the deliverable nodes (`Dxx`) on the PFD. The fixed columns are `ID` and `Description`; the rest are ExtraHeaders.

A typical column layout:

```tsv
ID	Description	Available Time	Max Revision
```

| Column name (English / Japanese) | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string  | Deliverable ID (corresponds to `Dxx` in the PFD) | ✅ | Fixed column |
| `Description` | string  | Deliverable name | ✅ | Fixed column |
| `Available Time` / `利用可能時刻` | float64  | The time at which the initial deliverable becomes available | ❌ | Effective only for deliverables with no upstream process. Empty/0 means "available from the start" |
| `Max Revision` / `最大版` | int  | The termination condition of the feedback loop | ❌ | Set only for feedback-source deliverables (integer ≥ 1). The loop terminates when revision ≥ Max Revision. For non-feedback deliverables, leave it empty or `-` |

### 5. Resource table

This is the dictionary of resource IDs referenced by `Needed Resources` in the atomic process table.

```tsv
ID	Description
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Resource ID (referenced by `Needed Resources`) | ✅ | Fixed column |
| `Description` | string | Description of the resource (such as a person or team name) | ✅ | Fixed column |

### 6. Composite deliverable table

This is a TSV that defines deliverable sets that bundle multiple atomic deliverables.

```tsv
ID	Description	Deliverables
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Composite deliverable ID | ✅ | Fixed column |
| `Description` | string | Description of the composite deliverable | ✅ | Fixed column |
| `Deliverables` | string | A comma-separated list of the contained atomic deliverable IDs (e.g., `D1, D2, D3`) | ✅ | Fixed column |

### 7. Composite process table

This is a TSV that defines higher-level processes that bundle multiple processes.

```tsv
ID	Description
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Composite process ID | ✅ | Fixed column |
| `Description` | string | Description of the composite process | ✅ | Fixed column |

### 8. Milestone table

This is a TSV that defines the ordering relationships between milestones. It is used when generating a master schedule with the `planmaster` command.

```tsv
ID	Description	Groups	Successors
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Milestone ID | ✅ | Fixed column |
| `Description` | string | Description of the milestone | ✅ | Fixed column |
| `Groups` | string | A comma-separated list of the group IDs belonging to this milestone | ✅ | Fixed column |
| `Successors` | string | A comma-separated list of successor milestone IDs | ✅ | Fixed column |

### 9. Group table

This is the dictionary of group IDs. It is referenced by the `Group` column of the atomic process table and the `Groups` column of the milestone table. It is used by the `planmaster` command.

```tsv
ID	Description
```

| Column name | Type | Meaning | Required | Notes |
| --- | --- | --- | --- | --- |
| `ID` | string | Group ID | ✅ | Fixed column |
| `Description` | string | Description of the group | ✅ | Fixed column |

### 10. Critical path analysis result

This is the output of the `criticalpath` command. It is not an input definition file.

```tsv
ATOMIC_PROCESS	TOTAL_FLOAT	MINIMUM_ELASTICITY
```

| Column name | Type | Meaning |
| --- | --- | --- |
| `ATOMIC_PROCESS` | string | Atomic process ID |
| `TOTAL_FLOAT` | float64 (≧ 0) | Total float (a process with `0.00` is a candidate on the critical path) |
| `MINIMUM_ELASTICITY` | float64 (≧ 0) or `"-"` | A metric for how much that process can stretch before it affects the overall deadline (`-` means no value) |

### Creation steps

These are the steps for creating a full set of PFD files for a new project.

#### Step 1: Draw the PFD

Create the PFD in draw.io.

- Draw processes as ellipses and deliverables as rectangles
- Draw atomic elements with thin lines (`strokeWidth=1`) and composite elements with thick lines (`strokeWidth>1`)
- Make edges that represent rework dashed
- Write the ID and description in the label, like `P1: Design`, `D1: Design document`

If there are nodes whose IDs have not been assigned, you can auto-number them with `pfdrenum`:

```console
$ pfdrenum -inplace path/to/pfd.drawio
```

#### Step 2: Generate templates for the TSV tables

Use the `pfdtable` command to generate a template for each table from the PFD:

```console
$ pfdtable -t cd -p path/to/pfd.drawio > cd.tsv
$ pfdtable -t ap -p path/to/pfd.drawio > ap.tsv
$ pfdtable -t ad -p path/to/pfd.drawio > ad.tsv
$ pfdtable -t r  -p path/to/pfd.drawio > r.tsv
```

If you have updated the PFD, you can update the tables while preserving the existing ones with the `-existing` option:

```console
$ pfdtable -t ap -existing path/to/ap.tsv -p path/to/pfd.drawio > ap_new.tsv
```

#### Step 3: Fill in the TSV tables

Add the necessary columns to the generated templates and fill in the values.

- **Atomic process table** (`ap.tsv`): Add the `Est. Work Volume`, `Est. Rework Volume Ratio`, and `Needed Resources` columns, and fill in the estimate values for each process. Add the `Start Condition`, `Milestone`, and `Group` columns as needed
- **Deliverable table** (`ad.tsv`): Add the `Available Time` and `Max Revision` columns, and fill in the attributes for the necessary deliverables
- **Resource table** (`r.tsv`): Fill in the IDs and descriptions of the resources (teams/assignees) used in the project

#### Step 4: Create the configuration JSON

Create a configuration JSON that bundles the paths to each file:

```json
{
  "pfd": "pfd.drawio",
  "atomic_process_table": "ap.tsv",
  "atomic_deliverable_table": "ad.tsv",
  "resource_table": "r.tsv",
  "composite_deliverable_table": "cd.tsv"
}
```

#### Step 5: Run the static check

Verify the consistency of the PFD and each table with `pfdlint`:

```console
$ pfdlint -f ./project.json
```

Fix the PFD and TSV until there are no more errors.

#### Step 6: Generate the schedule

Generate an execution plan with `pfdplan`:

```console
$ pfdplan -f ./project.json -poor -start 2025-06-01 -start-time 10:00 -duration 9 \
    -not-biz-days <(holidays -locale ja)
```

#### Step 7: Analyze the critical path

Identify bottlenecks with `criticalpath` (besides `-poor`, there are also `-best` and `-better`, but they usually never finish computing, so we recommend using `-poor`):

```console
$ criticalpath -f ./project.json -poor
```


pfdlint
-------
Detects problems in PFD notation.

### Usage
```console
$ pfdlint -h
Usage: pfdlint [options] -p <pfd> [-a <atomic-process-table>] [-d <deliverable-table>] [-c <composite-process-table>] [-r <resource-table>]

Options
  -ap string
    	path to the atomic process fsmtable
  -at string
    	path to the atomic deliverable fsmtable
  -atomic-deliverable string
    	path to the atomic deliverable fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -composite-process string
    	path to the composite process fsmtable
  -config string
    	path to the run config file
  -cp string
    	path to the composite process fsmtable
  -debug
    	debug mode
  -f string
    	path to the run config file
  -format string
    	format of the fsmreporter (default "tsv")
  -g string
    	path to the group table
  -group string
    	path to the group table
  -locale string
    	locale of the fsmreporter (default "ja")
  -m string
    	path to the milestone table
  -milestone string
    	path to the milestone table
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -r string
    	path to the resource fsmtable
  -resource string
    	path to the resource fsmtable
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ pfdlint -p ./pfd/encoding/drawio/testdata/example.drawio
  WARNING no-desc Please add a concise description.  [D2]
  ERROR   single-src      A deliverable is being output from multiple processes. A deliverable should be output from only one process.   [D3]

  $ pfdlint -locale en -p ./pfd/encoding/drawio/testdata/example.drawio
  WARNING no-desc Please add a concise description.       [D3]
  ERROR   single-src      A deliverable should be output from only one process. This includes output through feedback edges.      [D2]

  $ # Using config file instead of individual table specifications
  $ pfdlint -f ./config.json
  WARNING no-desc Please add a concise description.  [D2]
  ERROR   single-src      A deliverable is being output from multiple processes. A deliverable should be output from only one process.   [D3]
```


pfdtable
--------
Creates element tables from PFD. Updating element tables is also possible.

### Usage
```console
Usage: pfdtable [options]

Options
  -ap string
    	path to the atomic process fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -debug
    	debug mode
  -existing string
    	path of the existing fsmtable
  -i string
    	format of the input PFD
  -inplace
    	overwrite the file in place
  -input-format string
    	format of the input PFD
  -locale string
    	locale of the fsmreporter (default "ja")
  -o string
    	format of the output fsmtable
  -output-format string
    	format of the output fsmtable
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -silent
    	silent mode
  -t string
    	type of the table (available: ap(atomic-process), ad(atomic-deliverable), cp(composite-process), cd(composite-deliverable), r(resource))
  -type string
    	type of the table (available: ap(atomic-process), ad(atomic-deliverable), cp(composite-process), cd(composite-deliverable), r(resource))
  -v	show version
  -version
    	show version

Example
  $ pfdtable -t ad -p path/to/pfd.drawio
  ID      Description     Location
  D1      Implementation  https://example.com/1
  ...

  $ pfdtable -t ap -p path/to/pfd.drawio
  ID      Description
  P1      Implement
  ...

  $ pfdtable -t cp -p path/to/pfd.drawio
  ID      Description
  P1      Implement
  ...

  $ # Copy to clipboard as RTF (it is useful for pasting into Confluence and Microsoft Word and so on)
  $ pfdtable -t ad -o html path/to/pfd.drawio | textutil -stdin -format html -convert rtf -inputencoding UTF-8 -stdout | pbcopy

  $ # Print updated fsmtable from the existing fsmtable
  $ pfdtable -t ad -existing path/to/existing.tsv -p path/to/pfd.drawio
  ID      Description     Location
  D1      Implementation  https://example.com/1
  ...
```


pfdrenum
--------
Numbers PFD elements. Existing IDs are maintained.

### Usage
```console
$ pfdrenum -h
Usage: pfdrenum [options]

Options
  -debug
    	debug mode
  -inplace
    	overwrite the file in place
  -locale string
    	locale of the fsmreporter (default "ja")
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ pfdrenum path/to/pfd.drawio
  <mxfile host="65bd71144e">
    <diagram id="1ni4HEU6g7zc3-6eLzPC" name="P0">
    ...

  $ pfdrenum -inplace path/to/pfd.drawio
```


pfdplan
-------
Searches for optimal execution plans (Gantt charts) from PFD and environment.

### Usage
```console
$ pfdplan -h
Usage: pfdplan [-debug|-silent] -p <pfd> -a <atomic-process-table> -r <resource-table> -d <deliverable-table> [-start-time <start-time> -duration <duration> [-weekdays <weekdays>] [-not-biz-days <not-biz-days>]|-o plan-json|timeline-json|google-spreadsheet-tsv]

Options
  -ap string
    	path to the atomic process fsmtable
  -at string
    	path to the atomic deliverable fsmtable
  -atomic-deliverable string
    	path to the atomic deliverable fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -best
    	search best plan
  -better
    	search better plan
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -config string
    	path to the run config file
  -debug
    	debug mode
  -duration float
    	duration (default 9)
  -f string
    	path to the run config file
  -g string
    	path to the group table
  -group string
    	path to the group table
  -locale string
    	locale of the fsmreporter (default "ja")
  -m string
    	path to the milestone table
  -max-results int
    	upper bound of the number of results to return >= 1 (default 3)
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
  -milestone string
    	path to the milestone table
  -node-budget int
    	upper bound of the number of nodes to expand >= 1 (default 10000)
  -not-biz-days string
    	not business days except weekdays (comma separated dates. e.g. 2025-01-01,2025-01-02)
  -out-dir string
    	output directory
  -out-format string
    	output format (available: google-spreadsheet-tsv, plan-json, timeline-json)
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -poor
    	search plan by greedy algorithm (faster than best and better)
  -quality string
    	quality preset (available: s, m, l, xl, xxl) (default "small")
  -r string
    	path to the resource fsmtable
  -random-seed int
    	random seed (default 922990587439306466)
  -resource string
    	path to the resource fsmtable
  -restarts int
    	number >= 0 of restarts for diversity
  -silent
    	silent mode
  -start string
    	start day
  -start-time string
    	start time (default "10:00")
  -top-k-per-state int
    	upper bound of the number of transitions to consider per state >= 1 (default 128)
  -v	show version
  -version
    	show version
  -weekdays string
    	comma separated weekdays (available: sun,mon,tue,wed,thu,fri,sat) (default "mon,tue,wed,thu,fri")
  -weight float
    	weight >= 1.0 of Weighted A*. closer to 1.0 means closer to A*, greater than 1.0 means closer to greedy (default 2)

Example
    $ pfdplan -p path/to/pfd.drawio -a path/to/atomic_proc.tsv -d path/to/deliv.tsv -r path/to/resource.tsv -start-time 10:00 -duration 9 -not-biz-days <(holidays -locale ja)
    AtomicProcess[NumOfComplete]     StartTime       EndTime
    P1[1]   2025-10-04T00:00:00+09:00       2025-10-11T04:30:00+09:00
    P1[2]   2025-10-04T04:30:00+09:00       2025-10-11T06:45:00+09:00
    P1[3]   2025-10-04T06:45:00+09:00       2025-10-11T06:45:00+09:00
	...
```


pfddiff
-------
Compares two PFDs.

### Usage
```console
$ pfddiff -h
Usage: pfddiff [options] -p1 <pfd-a> -cd1 <composite-deliverable-table-a> -p2 <pfd-b> -cd2 <composite-deliverable-table-b>

Options
  -cd1 string
    	path to the composite deliverable fsmtable for PFD A
  -cd2 string
    	path to the composite deliverable fsmtable for PFD B
  -composite-deliverable1 string
    	path to the composite deliverable fsmtable for PFD A
  -composite-deliverable2 string
    	path to the composite deliverable fsmtable for PFD B
  -debug
    	debug mode
  -format string
    	format of the output (default "diff")
  -locale string
    	locale of the fsmreporter (default "ja")
  -p1 string
    	path to PFD A
  -p2 string
    	path to PFD B
  -pfd1 string
    	path to PFD A
  -pfd2 string
    	path to PFD B
  -prompt
    	AI-friendly output format
  -show-same
    	show same nodes and edges
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ pfddiff -p1 path/to/a.drawio -cd1 path/to/a_composite_deliv.tsv -p2 path/to/b.drawio -cd2 path/to/b_composite_deliv.tsv
  + P1 ----> D1
  - P2 ----> D2

  $ pfddiff -prompt -p1 path/to/a.drawio -cd1 path/to/a_composite_deliv.tsv -p2 path/to/b.drawio -cd2 path/to/b_composite_deliv.tsv
```


pfdquery
--------
Queries information about PFD elements.

### Usage
```console
$ pfdquery -h
Usage: pfdquery [options] -p <pfd> [-a <atomic-process-table>] [-d <deliverable-table>] [-r <resource-table>]

Options
  -ap string
    	path to the atomic process fsmtable
  -at string
    	path to the atomic deliverable fsmtable
  -atomic-deliverable string
    	path to the atomic deliverable fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -backward-reachable
    	backward reachable
  -backward-reachable-fb
    	backward reachable feedback destination
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -composite-process string
    	path to the composite process fsmtable
  -config string
    	path to the run config file
  -cp string
    	path to the composite process fsmtable
  -debug
    	debug mode
  -f string
    	path to the run config file
  -g string
    	path to the group table
  -group string
    	path to the group table
  -locale string
    	locale of the fsmreporter (default "ja")
  -m string
    	path to the milestone table
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
  -milestone string
    	path to the milestone table
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -r string
    	path to the resource fsmtable
  -reachable
    	reachable
  -resource string
    	path to the resource fsmtable
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ pfdquery -p path/to/pfd.drawio -reachable
  $ pfdquery -p path/to/pfd.drawio -backward-reachable
```


bizday
------
Calculates which business day a specified time corresponds to within business hours.

### Usage
```console
$ bizday -h
Usage: bizday [-start <start-day>] [-start-time <start-time>] [-duration <duration>] [-weekdays <weekdays>] [-not-biz-days <not-biz-days>] -time <time>

Options
  -debug
    	debug mode
  -duration float
    	duration (default 9)
  -locale string
    	locale of the fsmreporter (default "ja")
  -not-biz-days string
    	not business days except weekdays (comma separated dates. e.g. 2025-01-01,2025-01-02)
  -silent
    	silent mode
  -start string
    	start day
  -start-time string
    	start time (default "10:00")
  -t string
    	time
  -time string
    	time
  -v	show version
  -version
    	show version
  -weekdays string
    	comma separated weekdays (available: sun,mon,tue,wed,thu,fri,sat) (default "mon,tue,wed,thu,fri")

Example
  $ bizday -time '2016-01-02 15:00'
  1234

  $ bizday -start 2025-01-01 -start-time 10:00 -duration 9 -not-biz-days <(holidays -locale ja) -weekdays mon,tue,wed,thu,fri -time '2025-01-01 10:00'
  1234
```


planmaster
----------
Creates group-specific master schedules from execution plans and milestone/group information.

### Usage
```console
$ planmaster -h
Usage: planmaster [options] -f <project> <plan>

Options
  -ap string
    	path to the atomic process fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -b float
    	buffer multiplier (default 1)
  -buffer float
    	buffer multiplier (default 1)
  -config string
    	path to the run config file
  -debug
    	debug mode
  -f string
    	path to the run config file
  -g string
    	path to the group table
  -group string
    	path to the group table
  -locale string
    	locale of the fsmreporter (default "ja")
  -m string
    	path to the milestone table
  -milestone string
    	path to the milestone table
  -out-format string
    	output format (available: google-spreadsheet-tsv, plan-json, timeline-json)
  -p string
    	path to the plan
  -plan string
    	path to the plan
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ planmaster -f path/to/config.json -m path/to/milestone.tsv -g path/to/group.tsv path/to/plan.json
```


plantimeline
------------
Generates timeline data in specified format (such as Google Sheets Timeline) from execution plan JSON files.

### Usage
```console
$ plantimeline -h
Usage: plantimeline [options] -f <project> <plan>

Options
  -ap string
    	path to the atomic process fsmtable
  -at string
    	path to the atomic deliverable fsmtable
  -atomic-deliverable string
    	path to the atomic deliverable fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -config string
    	path to the run config file
  -debug
    	debug mode
  -duration float
    	duration (default 9)
  -f string
    	path to the run config file
  -locale string
    	locale of the fsmreporter (default "ja")
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
  -not-biz-days string
    	not business days except weekdays (comma separated dates. e.g. 2025-01-01,2025-01-02)
  -out-format string
    	output format (available: google-spreadsheet-tsv, plan-json, timeline-json)
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -r string
    	path to the resource fsmtable
  -resource string
    	path to the resource fsmtable
  -silent
    	silent mode
  -start string
    	start day
  -start-time string
    	start time (default "10:00")
  -v	show version
  -version
    	show version
  -weekdays string
    	comma separated weekdays (available: sun,mon,tue,wed,thu,fri,sat) (default "mon,tue,wed,thu,fri")

Example
  $ plantimeline -f path/to/config.json path/to/plan.json

  $ plantimeline -out-format timeline-json -f path/to/config.json path/to/plan.json
```


holidays
--------
Outputs a list of dates not considered business days in CSV format. Obvious dates such as Saturdays and Sundays in Japan are not output.

### Usage
```console
$ holidays -h
Usage: holidays -locale <locale>

Options
  -debug
    	debug mode
  -locale string
    	locale of the fsmreporter (default "ja")
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ holidays -locale ja
  2025-01-01
  2025-01-02
  2025-01-03
  ...
```


criticalpath
------------
Detects critical paths from PFD. Identifies the most time-consuming routes in a project to help optimize scheduling.

### Usage
```console
$ criticalpath -h
Usage: criticalpath [options] <pfd>

Options
  -ap string
    	path to the atomic process fsmtable
  -at string
    	path to the atomic deliverable fsmtable
  -atomic-deliverable string
    	path to the atomic deliverable fsmtable
  -atomic-process string
    	path to the atomic process fsmtable
  -best
    	search best plan
  -better
    	search better plan
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -config string
    	path to the run config file
  -debug
    	debug mode
  -f string
    	path to the run config file
  -g string
    	path to the group table
  -group string
    	path to the group table
  -locale string
    	locale of the fsmreporter (default "ja")
  -m string
    	path to the milestone table
  -max-results int
    	upper bound of the number of results to return >= 1 (default 3)
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
  -milestone string
    	path to the milestone table
  -node-budget int
    	upper bound of the number of nodes to expand >= 1 (default 10000)
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -poor
    	search plan by greedy algorithm (faster than best and better)
  -quality string
    	quality preset (available: s, m, l, xl, xxl) (default "small")
  -r string
    	path to the resource fsmtable
  -random-seed int
    	random seed
  -resource string
    	path to the resource fsmtable
  -restarts int
    	number >= 0 of restarts for diversity
  -silent
    	silent mode
  -top-k-per-state int
    	upper bound of the number of transitions to consider per state >= 1 (default 128)
  -v	show version
  -version
    	show version
  -weight float
    	weight >= 1.0 of Weighted A*. closer to 1.0 means closer to A*, greater than 1.0 means closer to greedy (default 2)

Example
  $ criticalpath path/to/pfd.drawio -a path/to/atomic_proc.tsv -r path/to/resource.tsv
```
