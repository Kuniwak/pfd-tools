Debugging Commands
==================

pfdrun
------
Simulates PFD execution. This is a debugging command.

### Usage
```console
$ pfdrun -h
Usage: pfdrun [options] [-p <pfd> [-a <atomic-process-table>] [-d <deliverable-table>] [-r <resource-table>]] [-f <config>] [-plan <plan>]

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
  -f string
    	path to the run config file
  -locale string
    	locale of the fsmreporter (default "ja")
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -plan string
    	path to the plan
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
    $ pfdrun -f path/to/config.json

    $ pfdrun -p path/to/pfd.drawio -a path/to/atomic_proc.tsv -d path/to/deliv.tsv -r path/to/resource.tsv

    $ pfdrun -f path/to/config.json -plan path/to/plan.json
```


pfdrungraph
-----------
Visualizes PFD execution graphs. Analyzes process dependencies and execution order to output graphical representations. This is a debugging command.

### Usage
```console
$ pfdrungraph -h
Usage: pfdrungraph [options] -p <pfd> [-a <atomic-process-table>] [-d <deliverable-table>] [-r <resource-table>]

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
  -locale string
    	locale of the fsmreporter (default "ja")
  -max-depth int
    	max depth (default 100)
  -max-results int
    	upper bound of the number of results to return >= 1 (default 3)
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
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
    	random seed (default 8746073225951359248)
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
  $ pfdrungraph -p path/to/pfd.drawio -a path/to/atomic_proc.tsv -d path/to/deliv.tsv -r path/to/resource.tsv
```


pfddot
------
Outputs PFD in Graphviz DOT format. This is a debugging command.

### Usage
```console
$ pfddot -h
Usage: pfddot [options] -p <pfd> -cd <composite-deliverable-table>

Options
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -debug
    	debug mode
  -locale string
    	locale of the fsmreporter (default "ja")
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ pfddot -p path/to/pfd.drawio -cd path/to/composite_deliverable.tsv
  digraph G {
    P1 [label="Process 1"];
    D1 [label="Deliverable 1"];
    P1 -> D1;
    ...
  }

  $ pfddot -p path/to/pfd.drawio -cd path/to/composite_deliverable.tsv | dot -Tpng -o output.png
```


pfddeadlock
-----------
Detects deadlocks in PFD. This is a debugging command.

### Usage
```console
$ pfddeadlock -h
Usage: pfddeadlock [options] -p <pfd> [-a <atomic-process-table>] [-d <deliverable-table>] [-r <resource-table>]

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
  -f string
    	path to the run config file
  -locale string
    	locale of the fsmreporter (default "ja")
  -maximal-available-allocations-threshold int
    	use only maximal available allocations if number of newly allocatable atomic processes is greater than the threshold. do not use maximal available allocations if threshold is not positive (default 10)
  -o string
    	output directory
  -out string
    	output directory
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
  $ pfddeadlock -p path/to/pfd.drawio
```


pfdparse
--------
Parses PFD. This is a debugging command.

### Usage
```console
$ pfdparse -h
Usage: pfdparse [options] -p <pfd> -cd <composite-deliverable-table>

Options
  -cd string
    	path to the composite deliverable fsmtable
  -composite-deliverable string
    	path to the composite deliverable fsmtable
  -debug
    	debug mode
  -locale string
    	locale of the fsmreporter (default "ja")
  -p string
    	path to the PFD
  -pfd string
    	path to the PFD
  -silent
    	silent mode
  -v	show version
  -version
    	show version

Example
  $ pfdparse -p path/to/example.drawio -cd path/to/composite_deliverable.tsv
  {
    "nodes": [
      {
        "id": "D1",
        "desc": "",
        "type": "DELIVERABLE"
      },
      ...
    ],
    "edges": [
      {
        "source": "P1",
        "target": "D1"
      },
      ...
    ],
	"composition": {
      "P0": [
        "P1",
        "P2"
      ],
	  ...
	]
  }
```
