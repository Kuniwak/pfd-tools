testdata/large_nofb
===================

A fixture that is `testdata/large` **with the rework (feedback) edges removed**.
It is used to run the execution model that does not handle feedback edges (`"feedback_mode": "disabled"`) on a large PFD.

Relationship to `testdata/large`
--------------------------------

`pfd.drawio` is `testdata/large/pfd.drawio` with the following removed:

- 6 dashed (`dashed=1`) edges: `D3 -> P6`, `D3 -> P10`, `D4 -> P11`, `D5 -> P2`, `D7 -> P6`, `D8 -> P7`
- Composite deliverable `D3`: it had only the edges above, so removing them makes it an isolated node with no edges (which would produce a `weak-conn` warning). It has also been removed from `cd.tsv`

That the node set and the solid-edge set match `testdata/large` is pinned by
`TestFixture_LargeNoFeedbackIsLargeWithoutFeedbackEdges` in `tools/pfdplan/cmd/cmd_test.go`.
If you change the `testdata/large` diagram, this test will fail; in that case, recreate the diagram in this directory as well.

The tables are those of `testdata/large` with columns dropped as follows.

| File | Dropped columns | Reason |
|---|---|---|
| `ad.tsv` | `Max Revision` | Because there are no longer any feedback-source deliverables. `-fb disabled` does not read this column |
| `ap_finite.tsv` | `Est. Rework Volume Ratio` | Same as above (rework volume is not used) |
| `ap_infinite.tsv` | `Est. Rework Volume Ratio`, `Needed Resources` | In addition to the above, `-res infinite` does not read needed resources |

`\complete(*)` in `Start Condition` is kept as-is. Even without feedback edges, it still means
"none of the atomic processes reachable backward are executable".

Intentionally kept
------------------

`D4`, `D5`, `D7`, `D8`, `D3.1`, `D3.2`, and `D3.3` are terminal deliverables that are produced but consumed by no one.
This is fine per the specification (`pfdlint` is clean in all four configurations), and it is a shape that arises naturally from removing the rework edges,
so they are kept as-is without adding consuming edges or deleting the deliverables.

Settings files
--------------

| File | Execution model |
|---|---|
| `project_finite.json` | Finite resources, no rework |
| `project_infinite.json` | Infinite resources, no rework |

The two patterns with rework are covered by `testdata/large/project.json` (finite resources) and
`testdata/large/project_infinite.json` (infinite resources).
