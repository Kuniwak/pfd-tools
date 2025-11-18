pfd-tools
=========
Tools related to Process Flow Diagram (PFD) such as static analyzers and schedulers.

**:bee: This repository is not accepting pull requests.**


Installation
------------
Download the latest binary from [Releases](https://github.com/Kuniwak/pfd-tools/releases) and place it in a directory that is in your PATH.


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
