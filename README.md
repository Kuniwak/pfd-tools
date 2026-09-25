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

This repository ships AI agent skills that automate the workflow described in [Input Files and Creation Steps](#input-files-and-creation-steps):

- **pjm-pfd-project-bootstrap** — scaffolds a new pfd-tools project from a `pfd.drawio.png` template and runs the pre-estimation lint gate.
- **pjm-pfd-estimation-plan** — produces a two-point (optimistic/pessimistic) estimate, analyzes the critical path, and generates a Mermaid Gantt or a Google Sheets Timeline.
- **pjm-pfd-replan** — reflects a progress table (`progress.tsv`) into the plan and re-plans when work is delayed.
- **pjm-pfd-delay-triage** — judges from the critical path's total float whether a delayed process affects the plan's end date and needs escalation.
- **pjm-pfd-migrate-v7** — migrates projects created by the pfd-tools 6.x or earlier skills to the 7.x format.

Each skill asks whether to work in English or Japanese before it starts.

### Claude Code

```console
$ claude
> /plugin marketplace add Kuniwak/pfd-tools
> /plugin install pfd-skills@pfd-tools
```

### Cursor

The skills live under `.cursor/skills/` (and `.agents/skills/`), so Cursor picks them up automatically when this repository is open. To use them in another project, copy the `pjm-pfd-*` directories from [`.ai/skills/`](.ai/skills) into that project's `.cursor/skills/`.

Tools
-----
For the options and usage examples of each tool, see `pfdhelp` (all tools), `pfdhelp <tool>` (a single tool), or each command's `-h`.
`pfdhelp -short` lists only the one-line description of every tool.

| Category | Command | Description |
| --- | --- | --- |
| Checking & formatting | [pfdlint](#pfdlint) | Detects notation problems in a PFD. |
| Checking & formatting | [pfdsort](#pfdsort) | Automatically arranges a PFD into a readable left-to-right layered layout. |
| Checking & formatting | [pfdfix](#pfdfix) | Connects unconnected edges that merely appear to touch a rectangle. |
| Checking & formatting | [pfdrenum](#pfdrenum) | Numbers the elements of a PFD. |
| Checking & formatting | [pfddupmark](#pfddupmark) | Adds ＊ to the IDs of duplicated deliverable displays. |
| Checking & formatting | [pfdcpcomp](#pfdcpcomp) | Creates skeletons of detail pages for composite processes. |
| Tables, estimates & tickets | [pfdtable](#pfdtable) | Creates and updates element tables from a PFD. |
| Tables, estimates & tickets | [pfdestcallout](#pfdestcallout) | Adds an estimation text box directly below each atomic process. |
| Tables, estimates & tickets | [pfdesttable](#pfdesttable) | Reads the values of the estimation boxes and outputs an estimate table (TSV). |
| Tables, estimates & tickets | [pfdres](#pfdres) | Generates a default value for the `必要資源` (needed resources) column of the atomic process table from a number of workers. |
| Tables, estimates & tickets | [pfdticket](#pfdticket) | Generates a ticket body for each atomic process. |
| Planning & scheduling | [pfdplan](#pfdplan) | Searches for an execution plan (Gantt chart) from a PFD and an environment. |
| Planning & scheduling | [criticalpath](#criticalpath) | Detects the critical path from a PFD. |
| Planning & scheduling | [planmaster](#planmaster) | Creates a master schedule by row and bar from an execution plan and classification tables. |
| Planning & scheduling | [plantimeline](#plantimeline) | Generates timeline data from an execution plan. |
| Planning & scheduling | [bizday](#bizday) | Calculates which business day a specified time falls on within business hours. |
| Planning & scheduling | [holidays](#holidays) | Outputs a list of dates that are not considered business days. |
| Query, comparison & help | [pfdquery](#pfdquery) | Queries information about PFD elements. |
| Query, comparison & help | [pfddiff](#pfddiff) | Compares two PFDs. |
| Query, comparison & help | [pfdhelp](#pfdhelp) | Displays the help of every command together. |
| Debugging | `pfdrun` | Simulates the execution of a PFD. |
| Debugging | `pfdrungraph` | Visualizes the execution graph of a PFD. |
| Debugging | `pfddot` | Outputs a PFD in Graphviz DOT format. |
| Debugging | `pfddeadlock` | Detects deadlocks in a PFD. |
| Debugging | `pfdparse` | Parses a PFD and outputs it as JSON. |


Execution Model
---------------
Schedulers such as `pfdplan` simulate the execution of a PFD as a state machine.
There are four execution models, formed by combining how resources are handled with how feedback edges (rework) are handled; you select one with the `-res` / `-fb` options (or the `resource_mode` / `feedback_mode` keys of the project configuration file).

| `-res` | `-fb` | Resources | Rework | Additional input required |
| --- | --- | --- | --- | --- |
| `infinite` (default) | `disabled` (default) | Resource contention is not considered | Rework is not considered | None |
| `infinite` (default) | `enabled` | Resource contention is not considered | Rework is considered | The `予想手戻り作業量割合` (expected rework ratio) column of the atomic process table and the `最大版` (maximum revision) column of the deliverable table |
| `finite` | `disabled` (default) | Resource contention is considered | Rework is not considered | The resource table (`-r`) and the `必要資源` (needed resources) column of the atomic process table |
| `finite` | `enabled` | Resource contention is considered | Rework is considered | All of the above |

- `-res infinite` does not read the resource table or the `必要資源` column (they are ignored even if specified; however, as long as they are written, they are still subject to `pfdlint`'s syntax checks). All processes can run simultaneously, and the lead time of each process is exactly its `予想作業量` (expected work volume).
- `-res finite` determines resource contention from the resource table and the `必要資源` column, and does not run processes that request the same resource at the same time.
- `-fb disabled` does not read the `予想手戻り作業量割合` or `最大版` columns (as with `-res infinite`, they are still subject to syntax checks as long as they are written). PFDs containing feedback edges are not accepted, and `pfdlint` reports them as `feedback-edge-not-available`.
- `-fb enabled` simulates rework caused by feedback edges. The rework volume is determined by `予想手戻り作業量割合`, and loop termination by `最大版`.

In the glossary ([docs/GLOSSARY.md](docs/GLOSSARY.md)), the infinite-resource model is called ISM (infinite-resource single-deliverable execution model) and the finite-resource model is called FSM (finite-resource single-deliverable execution model).

`-res` / `-fb` are common to every command that uses the execution model (`pfdlint`, `pfdplan`, `criticalpath`, `plantimeline`, `pfdquery`, `pfdtable`).
`pfdtable` does not generate columns or tables that the execution model does not read.


Input Files and Creation Steps
------------------------------
For the definitions of the input files and the steps to create them, see [docs/FILE_FORMAT.md](docs/FILE_FORMAT.md).


pfdlint
-------
Detects notation problems in a PFD.
For the invariants it checks, see [docs/INVARIANTS.md](docs/INVARIANTS.md).


pfdsort
-------
Automatically arranges a PFD (draw.io diagram) that has become messy, for example through machine generation, into a readable left-to-right layered layout.
Upstream processes are placed on the left and final deliverables on the right, and each page is arranged independently.

Apart from feedback edges (dashed lines), the input must be acyclic (invariant `acyclic-except-fb`).
The following diagrams are rejected with an error (check the violation itself with `pfdlint`):

- Diagrams that become cyclic when deliverables with the same ID are identified (a logical-level violation of invariant `acyclic-except-fb`)
- Diagrams in which deliverable cells with the same ID are connected by a solid line (violation of invariant `no-d2d`)
- Diagrams containing a feedback edge whose source is not a deliverable or whose target is not a process (violation of invariant `no-p2d-fb`)

Arrangement preserves the following:

- The width and height of each vertex (only positions are rewritten; the sole exception is duplicated display cells, which are aligned with the box they are folded into, as described below)
- The top edge (minimum Y) of the whole diagram
- Cells and edges on the comment layer (they are not moved)
- Estimation boxes directly below atomic processes (those placed by `pfdestcallout`; they follow the process to directly below its new position)
- The appearance of edges, such as thickness, color, dashes (`dashed=1`), and arrowheads

Arranged edges become straight lines connecting their endpoints (`edgeStyle=none`), and fixed connection points and waypoints are removed.
Only the target edges of arranged pages (those whose both ends are deliverables, processes, or connectors) are rewritten; edges whose endpoints are text boxes or the like are left untouched.

Edges that span a large distance horizontally (rank difference) or vertically (row difference) are shortened by placing a duplicated display of the deliverable near the process.
The thresholds are `-dup-rank-span` (default 1) and `-dup-row-span` (default 5); an edge is duplicated only when its rank difference or row difference exceeds these values.
Larger values suppress duplication, and making both large enough disables it.
Duplicated displays already in the diagram (including hand-drawn ones) are folded back into a single box every time before duplication is re-planned.
Deliverables that no longer need duplication return to a single box.
The label, style, and size of duplicated cells are aligned with the box they are folded into (if you gave a duplicated cell its own color, that color is lost).
The IDs of duplicated cells get ＊ (see [pfddupmark](#pfddupmark)).
After arrangement, boxes of the same deliverable may be split into a "box with only producing edges" and a "box with only consuming edges"; if they are in different columns, that is the intended result.

Repeated arrangement settles within a few runs, and running it again on a settled diagram does not change the result.
However, it does not necessarily settle in a single run.

With `-break-cycles`, pages that remain cyclic even after excluding feedback edges (such as context diagrams in which composite processes depend on each other) are also arranged (the default is off, and cycles are errors).
Edges on a cycle (solid lines from a deliverable to a process) are cut for layout purposes only, and each cut edge is represented by a duplicated display of the deliverable.
No edge is deleted or turned into `dashed=1`, so the logical PFD of the output is identical to the input (`pfdlint` results do not change and `pfddiff` reports no differences).
Cycles that contain no solid line from a deliverable to a process (diagrams that are not valid PFDs, such as those with edges between processes) cannot be cut and result in an error.

Passing comma-separated page names to `-only-page` arranges only those pages and leaves the other pages as they are (e.g. `-only-page P123,P456`).
A page name is the composite process ID (e.g. `P123`) for a detail page, or `P0` for the context diagram (which also matches the locale-dependent actual page names `Page-1`/`ページ-1`).
Matching is done on the ID part of the page name, so a page name with a description (`P123: foo bar`) can be selected by its ID alone.
Specify page names whose description contains a comma by ID only.
A warning is issued if a specified page name does not exist anywhere.

Passing comma-separated node IDs to `-only-node` repositions only those nodes and moves the other nodes as little as possible (e.g. `-only-node P1,D2`).
Use this when you want to tidy just a few added or moved nodes while preserving an existing layout that is close to hand-drawn.
This behaves differently from global arrangement: it does not perform ranking, rebuild duplicated displays, or reattach edges.
The new position of a specified node is determined locally from the current positions of the neighbors connected by solid lines (directly to the right of the producing process, directly to the left of the consuming process, and vertically at the average of the neighbors' centers).
IDs are matched against the label IDs of deliverables and processes (`D1`/`P123`, etc.), and as with `-only-page`, nodes can be selected by ID alone even if a description is written.
If a deliverable with the same ID is split into multiple cells by duplicated display, all cells are targeted, and each cell is placed separately according to its own connections.
Edge straightening is limited to edges connected to the moved nodes, and duplicated display marks (＊) are not reassigned.
A warning is issued if a specified ID does not match any node.
When `-break-cycles` is also given, cycles are cut only to determine the order in which the selected nodes are arranged, and the cut edges remain in the diagram.

`-pos-restriction` determines, with `-only-node`, how far unselected nodes are allowed to move (default `lock`).
`lock` does not move other nodes at all; if the place for a selected node is occupied, the selected node itself is moved down out of the way.
`free-v` allows only vertical movement of other nodes and `free-h` only horizontal movement, pushing other nodes that overlap a selected node aside along that axis by the minimum amount.
`free` pushes them in whichever direction, vertical or horizontal, requires the smaller movement.
It is ignored when `-only-node` is not given.

`-hgap` / `-vgap` adjust the gaps between columns and between rows.

Input and output are the same as `pfdfix`: it reads the file given as an argument (standard input if omitted) and by default writes the arranged draw.io to standard output.
With `-inplace`, it overwrites the file (it writes only after processing succeeds, so the original file is not corrupted on failure).


pfdfix
------
Connects edges that visually link processes and deliverables with arrows but have no `source`/`target` set, with endpoints that merely appear to touch a rectangle.
The bounding rectangle of each vertex is used as a hitbox, and the connection target is determined by which hitbox the endpoint of an edge lacking `source`/`target` falls into.

- The endpoint falls into exactly one hitbox: `source`/`target` is filled in and the redundant endpoint coordinates are removed (waypoints are kept).
- The endpoint falls into no hitbox (unhit): by default a WARN is emitted and the edge is left as is. With `-delete-unhit`, the edge is deleted.
- Source and target fall into the same hitbox (self-loop): not connected. By default a WARN is emitted and the edge is left as is; with `-delete-unhit`, the edge is deleted.
- The endpoint falls into multiple overlapping hitboxes and cannot be resolved (ambiguous): a WARN is emitted and the edge is left as is (unchanged).

`-hitbox-expand` scales hitboxes around their centers.
`x:y` means x times horizontally and y times vertically, and `n` means `n:n` (e.g. `0.9` shrinks to 90%, `1.1:0.9` means 110% horizontally and 90% vertically).


pfdrenum
--------
Numbers the elements of a PFD.
Existing IDs are kept, and labels are normalized to the form `ID: description`.
Duplicated display marks (＊) are reapplied to the numbered labels (see [pfddupmark](#pfddupmark) for the marking convention).
When a composite process is numbered, the name of its detail page follows (`実装する` → `P7: 実装する`).
With `-inplace`, it overwrites the file (it writes only after numbering succeeds, so the original file is not corrupted on failure).

### Numbering a PFD split across multiple files

Deciding the numbering requires the whole PFD, but in workflows that combine pages with `drawiocat`, numbering the combined output does not leave the IDs in the original files.
Therefore, deciding and applying are separated.
`-out-format tsv` outputs only the numbering plan as TSV, and `-renum-plan` applies that plan to a single file.

```console
$ # 1. Decide: combine everything and output the plan
$ drawiocat *.drawio | pfdrenum -out-format tsv >renum.tsv
Key	ID
(実装する)	P3
(設計する)	P2
[設計書]	D3

$ # 2. Apply: write back to the original files one by one
$ for f in *.drawio; do pfdrenum -renum-plan renum.tsv -inplace "$f"; done

$ # 3. Verify: application is complete if there are 0 rows
$ drawiocat *.drawio | pfdrenum -out-format tsv
Key	ID
```

The plan's `Key` is the node's label before numbering, enclosed according to its kind (`(label)` for processes, `[label]` for deliverables), and `ID` is the newly assigned ID.
Rows contain only newly numbered nodes; already numbered nodes are not listed.
Because the key is the label, you do not need to know which page is in which file, and the result does not depend on the combination order of `drawiocat` or on reassignment of page ids.
Even when the same label appears in multiple files, as with boundary deliverables, the same ID is assigned.
A thin line and a thick line (atomic and composite) with the same label are treated as one element, and a mismatch is an error.

Applying with `-renum-plan` is a syntactic rewrite without normalization.
Therefore, it can also be applied to files that cannot be normalized on their own (such as files containing only detail pages without the body of the composite process).
Because it does not touch the labels and page names of already numbered nodes that are not in the plan, unlike the default mode (`pfdrenum path/to/pfd.drawio`), no normalization to `ID: description` is performed.

Missed applications can be detected not only by the verification above (0 rows) but also by `pfdlint`'s `no-desc` (because unnumbered nodes have an empty description).

The plan is validated when it is read.
It is an error if any of the following holds: the header is not `Key<TAB>ID`, an `ID` is not of the form `P%d` or `D%d`, a `Key` is an ID, a `Key` is not enclosed by its kind (old-format plan), the kind of a `Key` and the kind of its `ID` disagree, there are multiple rows with the same `Key`, or the same `ID` is assigned to multiple keys.
Regenerate old-format plans with `pfdrenum -out-format tsv`.
Even when combined with `-inplace`, the plan is read before the target file is opened, so a broken plan never causes the target file to be lost.

### Numbering across variations

Variations that replace the breakdown (detail pages) of a composite process cannot be combined into a single PFD (invariant `uniq-page-name`; see [docs/INVARIANTS.md](docs/INVARIANTS.md)).
If you number them one at a time, `P7` in variation A and `P7` in variation B become different elements.
Therefore, the numbering plan is used as an **ID ledger** shared by all variations.
Passing `-renum-plan` together with `-out-format tsv` outputs a new plan built on that plan.
Labels in the base reuse the same IDs, and new numbering continues from the base's maximum ID.
The output is base ∪ new, so the input and output have the same shape, and you can grow the ledger by passing the variations through one at a time.

The ledger only knows the IDs it assigned, so to avoid colliding with IDs that have already been assigned by hand, give a lower bound for numbering with `-min-id`.
Compute the lower bound for each file with `-out-format maxid` and concatenate the results (`-min-id` accepts multiple lines and takes the maximum for each kind).
If you stream multiple `.drawio.png` files together with `cat`, only the first one is read and the lower bound becomes too small.

```console
$ # 0. Compute the lower bound of existing IDs from all variations (compute per file and concatenate)
$ floor=$(for f in parent.drawio variants/*/*.drawio; do pfdrenum -out-format maxid "$f"; done)
$ echo "$floor"
P1	D2
P0	D2
P0	D2

$ # 1. Decide: start from an empty ledger and grow it by passing variations through one at a time
$ printf 'Key\tID\n' >renum.tsv
$ for v in variants/*; do
>   drawiocat parent.drawio "$v"/*.drawio |
>     pfdrenum -min-id "$floor" -renum-plan renum.tsv -out-format tsv >renum.tsv.new
>   mv renum.tsv.new renum.tsv
> done
$ cat renum.tsv
Key	ID
(実装する)	P3
(設計する)	P2
(調査する)	P4
[設計書]	D3
[調査結果]	D4

$ # 2. Apply: write back to all files
$ for f in parent.drawio variants/*/*.drawio; do pfdrenum -renum-plan renum.tsv -inplace "$f"; done

$ # 3. Verify: application is complete if every variation yields 0 rows
$ for v in variants/*; do drawiocat parent.drawio "$v"/*.drawio | pfdrenum -out-format tsv; done
```

The folding in step 1 is pinned by `TestMainCommandByArgs_VariantsUniqueNumbering` in `tools/pfdrenum/cmd/cmd_test.go`, and the fact that passing the variations together is an error is pinned by `TestMainCommandByArgs_MergedVariantsAreRejected` in the same file.
This console block is not executed, so when you change the procedure, also update the corresponding tests.

Labels common across variations (boundary deliverables from the parent diagram, or processes that appear in every variation) get the same ID in all variations through reuse of the ledger.
The folding is idempotent: running a second pass with the same ledger does not change it.

`-min-id` accepts a sequence of IDs separated by whitespace or commas (e.g. `P12 D34`) and raises the lower bound of numbering for processes and deliverables respectively.
The output of `-out-format maxid` can be passed as is.
Combining it with modes that do not number (applying with `-renum-plan`, `-out-format maxid`) is an error.

ID collisions are detected as errors when an ID in the ledger collides with an already numbered element in the input, and when a ledger in which the same ID is assigned to multiple keys is read.
However, if "a variation has all of its elements already numbered and those IDs appear neither in the ledger nor in other variations", omitting step 0 (`-min-id`) leaves the collision unnoticed.
Do not omit step 0.


pfddupmark
----------
Adds ＊ to the IDs of duplicated deliverable displays so that you can tell from the diagram which ones are duplicates and which is the original (a duplicate of `D10: foo` becomes `D10＊: foo`).
`pfdsort` and `pfdrenum` output marks from the start, so you need to run this command separately only to fix diagrams in which duplicated displays were added by hand or the original was deleted.

The decision is made per page.
When a page has two or more cells with the same deliverable ID, one of them is the original and is left unmarked, and the rest are marked.
A cell that is the only one on its page is always unmarked, so stray marks left on a diagram whose duplicates were merged back into one are also removed.
When the same deliverable is drawn on both the parent diagram and a detail page, as with boundary deliverables of a composite process, both are unmarked if there is one on each page.

The original is "the cell that has an edge from the process that outputs the deliverable (a solid incoming edge, i.e. a producing edge)".
Incoming rework (dashed) edges are not counted as producing edges.
If no cell on the page has a producing edge (initial deliverables, or diagrams where the original was deleted and only duplicated displays remain), the cell with the smallest cell ID becomes the original.

Marks are written as the full-width ＊, but the half-width `*` is also accepted as a duplicate mark when reading.
Existing marks keep their character, and stacked marks are normalized to one.
The mark is placed on the ID side (the style `D10: foo＊` on the description side is not used).
Unnumbered labels (cells with only a description and no `D10:`) are not marked.
Unnumbered diagrams get marks once they are numbered with `pfdrenum`.

Labels with ＊ are read as `D10` by all other tools.
Therefore, adding marks does not change `pfdlint` results and does not produce differences in `pfddiff`.
A mark is stripped only when what remains after stripping is a deliverable ID, so an unnumbered deliverable named `foo＊` or `P3＊` does not turn into something else.

It is idempotent.
HTML labels (such as `<b>D10</b>: foo`) are also rewritten without breaking the tags.

Input and output are the same as `pfdsort`: it reads the file given as an argument (standard input if omitted) and by default writes to standard output.
With `-inplace`, it overwrites the file (it writes only after processing succeeds, so the original file is not corrupted on failure).
You can target only the pages passed to `-only-page`.


pfdcpcomp
---------
Creates skeletons of detail pages for composite processes (thick-lined ellipses).
For a composite process `Px`, it prepares a page in which the deliverables flowing into and out of `Px` in the parent diagram are arranged with inputs in a single left column and outputs in a single right column.
The skeleton assumes that a human draws the wiring in between (the internal processes and edges).
Inputs and outputs are determined, including feedback edges, as edges into `Px` being inputs and edges out of it being outputs.
In addition to supplying pages for composite processes without detail pages (those that trigger a `pfdlint` WARNING), it also targets composite processes that already have pages.

- If there is no page corresponding to `Px`: creates a new page `Px: description` with inputs in the left column and outputs in the right column. For an unnumbered composite process, the page name is the label itself.
- If `Px` has a page but boundary deliverables are missing: adds only the missing deliverables. Pages are matched on the ID part of the page name, and the page name is not changed. To make it obvious that they are not wired yet, the added deliverables are placed above the topmost node.

It is idempotent.
Applying it repeatedly to the same file does not change the result.


pfdtable
--------
Creates element tables from a PFD.
Passing existing element tables to `-existing` also allows updating them.
With `-inplace`, it overwrites the element table passed to `-existing` (it writes only after assembly succeeds, so the original file is not corrupted on failure).

Piping the output of `-out-format html` to `textutil -stdin -format html -convert rtf -inputencoding UTF-8 -stdout | pbcopy` copies it to the clipboard as RTF that can be pasted into Confluence or Microsoft Word.


pfdestcallout
-------------
So that two-point estimates (optimistic/pessimistic) can be entered graphically in draw.io, it adds an estimation text box (placeholder `楽観: d 悲観: d`, i.e. "optimistic: d pessimistic: d") directly below each atomic process.
`d` is in business days, and a human replaces it with the actual number of days, such as `2d`.

- The box is placed directly below the atomic process (touching the bottom edge of the process). Which process an estimate belongs to is expressed by adjacency.
- The box is added as a shape with the `text` style on the same layer as the process, and does not affect the judgments of `pfdlint` / `pfdplan` (it is not recognized as a process or deliverable).
- It is idempotent. Atomic processes that already have an estimation box directly below them are not changed, so values already entered are kept. Boxes are not added to composite processes, deliverables, or connectors.
- In addition to `.drawio`, it also accepts `.drawio.png` as input, and returns PNG output for PNG input.


pfdesttable
-----------
Reads the values of the estimation boxes placed by `pfdestcallout` (e.g. `楽観: 2d 悲観: 3d`) and outputs an estimate table (TSV) per atomic process to stdout.

- A column `<label>作業量` (e.g. `楽観作業量`, "optimistic work volume") is created for each label that appears in the boxes. If you write `楽観: 2d 最頻: 3d 悲観: 5d` in a box, a three-point estimate also becomes columns as is. However, boxes are identified by the same contract as `pfdestcallout` (a text shape whose value contains both `楽観` and `悲観`), so boxes that do not contain these two words are not recognized.
- Atomic processes without an estimation box, unfilled boxes (still the placeholder `d`), values that conflict between multiple boxes, and floating boxes that are not directly below any process produce a warning on stderr and leave the cell empty (empty cells are reliably turned into errors by the downstream `pfdplan`). Cells on the comment layer are not considered estimation boxes.
- The output is intended to be JOINed with the atomic process table (`ap.tsv`) using `qhs` or similar to derive per-scenario tables for `pfdplan` (e.g. `ap_optimistic.tsv`).
- In addition to `.drawio`, it also accepts `.drawio.png` as input (the output is always TSV).


pfdres
------
Generates a value usable in the `必要資源` (needed resources) column of the atomic process table from a number of workers.

The output is `<label>:1` separated by semicolons `;`.
Labels carry over as `A`, `B`, ..., `Z`, `AA`, `AB`, ....
`;` means OR (the planner picks one available person), so `pfdres <count>` is a default needed resource meaning "any one of that many people is in charge".
`-locale` accepts `ja` / `en`, but the output is the same.

It can be used as the default value for empty cells when deriving scenario tables with `qhs`:

```console
$ qhs -H -O -t -T "SELECT ID, ..., CASE WHEN LENGTH(\"楽観必要資源\") > 0 THEN \"楽観必要資源\" ELSE '$(pfdres 3)' END AS \"必要資源\", ... FROM \"-\"" < ap.tsv > ap_optimistic.tsv
```


pfdticket
---------
Generates a ticket body (summary and description) for each atomic process from a PFD and the deliverable table.
It does not depend on any ticket management system; it only outputs text.

- The summary has the form `P<n>: <process description>`.
- The description is Markdown consisting of `# 入力成果物の一覧` (list of input deliverables; non-feedback input deliverables) and `# 出力成果物の一覧と品質基準とレビューア` (list of output deliverables with quality criteria and reviewers; for each non-feedback output deliverable, `品質基準` (quality criteria) = the `レビュー基準` (review criteria) column of the deliverable table, `レビューア` (reviewer), and `成果物リンク` (deliverable link)).
- `成果物リンク` outputs the value of the `URL` column of the deliverable table if present, and otherwise a fixed placeholder (`ja`: `（完了時に URL を記入）` / `en`: `(fill in the URL when done)`).
- The output wording switches with `-locale` (`ja` / `en`, default `ja`). The column headings matched in the deliverable table default to names depending on `-locale` (`ja`: `フォーマット`/`レビュー基準`/`レビューア`/`URL`, `en`: `Format`/`Review Criteria`/`Reviewer`/`URL`), and can be specified individually with `-format-column` / `-review-criteria-column` / `-reviewer-column` / `-url-column`.


pfdplan
-------
Searches for an execution plan (Gantt chart) from a PFD and an environment.


criticalpath
------------
Detects the critical path from a PFD.


planmaster
----------
Creates a master schedule by row and bar from an execution plan and classification tables.
For the specification of the classification tables, drawing rules, and meta tables, see the Specification section of `planmaster -h`.


plantimeline
------------
Generates timeline data in a specified format (such as a Google Sheets Timeline) from an execution plan file in JSON format.

### Emphasizing bars (critical path, etc.)

Passing a TSV with an `ID` column (an emphasis ID table) to `-em-tsv` draws only the bars with those IDs emphasized (the `crit` tag for mermaid, a color for PlantUML, and an additional trailing `Emphasis` column for `google-spreadsheet-tsv`).
The decision of what to emphasize is outside the tool, so filtering the output of `criticalpath` gives you critical path emphasis.
The `ID` column is identified by name and other columns are ignored, so the filtered result can be passed as is.

```console
$ criticalpath -poor -f path/to/project.json >cp.tsv
$ qhs -H -O -t -T 'SELECT ID FROM cp.tsv WHERE "最大弾性値（全余裕）" < 0.0001' >em.tsv
$ plantimeline -f path/to/project.json -out-format mermaid -em-tsv em.tsv path/to/plan.json
```

`pfdplan` and `planmaster` accept the same option (for `planmaster`, the emphasis targets are milestone IDs).
IDs that do not exist in the target are warned about and ignored.
Emphasis is a visual specification, so specifying `-em-tsv` together with `-out-format plan-json` / `timeline-json` is an error.
It can also be specified with `emphasis_tsv_path` in the project configuration file, in which case it is simply not used for these output formats.
For details, see [Input file definitions and creation steps](docs/FILE_FORMAT.md#11-emphasis-id-table).


bizday
------
Calculates which business day a specified time corresponds to within business hours.


holidays
--------
Outputs a list of dates that are not considered business days, in CSV format.
Obvious dates such as Saturdays and Sundays in Japan are not output.

The output can be passed directly to `-not-biz-days` of `pfdplan` or `bizday` (e.g. `-not-biz-days <(holidays -locale ja)`).


pfdquery
--------
Queries information about PFD elements.


pfddiff
-------
Compares two PFDs.
Since it compares only nodes and edges, the PFDs do not need to be complete.

### Comparing fragments

When comparing on its own a file that contains only detail pages whose composite process bodies live in another file (a fragment for overlaying, which becomes a complete PFD only after being merged with a base PFD using `drawiocat`), add `-fragment`.
Without it, the composite process referenced by the page name is not in the diagram, so the file is rejected as a violation of invariant `page-name-comp` ([docs/INVARIANTS.md](docs/INVARIANTS.md)).

```console
$ pfddiff -fragment -p1 variant/incl-C.drawio.png -p2 variant/incl-C.drawio
```

`-fragment` is only for comparing fragments with each other and cannot be used to compare a fragment with a complete PFD.
Whether something is a composite process is determined by "thick line AND a detail page with the same name in the same file", so a thick-lined ellipse whose detail page is in the parent file is demoted to an atomic process in the fragment alone.
Between fragments, the same demotion happens on both sides so the diff is preserved, but if only one side is a complete PFD, even the same diagram produces differences in node kind.

`-fragment` relaxes only the page name check (`page-name-comp`).
If the element with the ID referenced by a page name is in the diagram and is not a composite process (a typo in the page name), it is an error even with `-fragment`.
Relaxed pages are reported in warnings.
`pfdlint` has no such relaxation.


pfdhelp
-------
Displays the help of every pfd-tools command together.
It collects the actual `-h` output of each tool as is, so you always get up-to-date help.
Passing a tool name displays only that tool.
With `-short`, it aggregates only the one-line description of each tool as TSV of the form `name<TAB>description`.
Every tool supports `--short-help` in common, which outputs only its one-line description.
