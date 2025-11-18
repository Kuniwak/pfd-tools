Data Invariant Conditions
============

Invariant Conditions that Graphs Must Satisfy
----------------------
| ID | Condition |
|:---|:-----------|
| in-field | All elements appearing at both ends of edges or feedback edges are included in the element set. |


Invariant Conditions that PFD Must Satisfy
---------------------
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
| consistent-input-comp | The input deliverable set of a composite process matches the set of input deliverables of atomic processes within the composite process that are not output deliverables of any atomic process within the composite process. |
| consistent-output-comp | The output deliverable set of a composite process is a subset of the union of output deliverables of atomic processes within the composite process. |