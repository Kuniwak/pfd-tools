Data Invariant Conditions
=========================
The ID of an invariant condition corresponds to the check named by replacing `-` with `_` (`pfd/pfdcheckers/<id>.go`) and its test (`<id>_test.go`). For example, `acyclic-except-fb` is `pfd/pfdcheckers/acyclic_except_fb.go`. However, the correspondence is not surjective: some checks in `pfd/pfdcheckers/` are not yet registered in this table.

Invariant Conditions that Graphs Must Satisfy
---------------------------------------------
| ID | Condition |
|:---|:-----------|
| in-field | All elements appearing at both ends of edges or feedback edges are included in the element set. |


Invariant Conditions that PFD Must Satisfy
------------------------------------------
| ID | Condition |
|:---|:-----------|
| graph-pfd | PFD is a graph. |
| consistent-desc | All elements with the same ID have the same description. |
| ex-input | Every process has at least one input deliverable. |
| ex-output | Every process has at least one output deliverable. |
| no-d2d | There are no edges directly connecting deliverable to deliverable. |
| no-p2p | There are no edges directly connecting process to process. |
| no-p2d-fb | Every feedback edge starts from a deliverable and ends at a process. |
| single-src | If any deliverable has a process that outputs it, it is unique. This includes output through feedback edges. |
| acyclic-except-fb | Removing feedback edges makes the graph acyclic. |
| weak-conn | Any final deliverable can be reached from any initial deliverable. |
| finite | Process set, deliverable set, and edge set are all finite sets. |
| disj-or-psubset-comp | Different composite processes are either disjoint or one is a proper subset of the other. |
| consistent-input-comp | The input deliverable set of a composite process matches the set of input deliverables of atomic processes within the composite process that are not output deliverables of any atomic process within the composite process. Composite deliverables are expanded into the atomic deliverables of their breakdown before comparison. |
| consistent-output-comp | The output deliverable set of a composite process is a subset of the union of output deliverables of atomic processes within the composite process. Composite deliverables are expanded into the atomic deliverables of their breakdown before comparison. |
| acyclic-cd-comp | A composite deliverable does not transitively contain itself. |


Invariant Conditions that Execution Models Must Satisfy
-------------------------------------------------------
| ID | Condition |
|:---|:-----------|
| feedback-edge-not-available | In an execution model that does not handle feedback edges (`-fb disabled`), the PFD has no feedback edges. |


Invariant Conditions that drawio Page Structure Must Satisfy
------------------------------------------------------------
These are the assumptions about page names and ids when a PFD is encoded as multiple drawio pages. Page names and ids do not remain in the PFD data, so they cannot be checked by the PFD invariant conditions above. Instead, `NormalizeDiagrams` in `pfd/pfdencoding/pfddrawio/normalize.go`, which decodes the encoding, checks them, and `TestNormalizeDiagramsError` in `normalize_test.go` pins them down with the ID included in the case name.

Among the assumptions checked by `NormalizeDiagrams`, the encoding of kinds by the enclosing shape and line width (such as drawing the same element ID as both a rectangle and an oval) and cycles in nested breakdowns are not yet registered. These are the cases without an ID in `TestNormalizeDiagramsError`.

| ID | Condition |
|:---|:-----------|
| page-name-comp | The name of a page that is not the context diagram matches the ID of an element drawn as a composite process somewhere in the PFD. Even when the page name includes a description (`P1: Design`), only the ID part is matched. |
| uniq-page-name | The composite process referred to by the name of a page that is not the context diagram is unique. That is, a composite process has one breakdown per PFD (variations of a breakdown cannot be combined into one PFD; see "Numbering across variations" in the README). |
| uniq-context-page | There is one context diagram page per PFD. |
| uniq-page-id | Page ids are unique. This also applies to the context diagram page. |
