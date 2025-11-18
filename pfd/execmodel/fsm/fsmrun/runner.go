package fsmrun

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/Kuniwak/pfd-tools/pairs"
	"github.com/Kuniwak/pfd-tools/pfd"
	"github.com/Kuniwak/pfd-tools/pfd/execmodel/fsm"
	"github.com/Kuniwak/pfd-tools/sets"
)

var ErrQuit = errors.New("quit")

type Runner struct {
	Reader   io.Reader
	Writer   io.Writer
	Env      *fsm.Env
	State    fsm.State
	History  []*fsm.Trans
	Commands []Command
}

type Command struct {
	Triggers    []string
	Usage       string
	Description string
	Action      func(r *Runner, args []string) error
}

func NewHelpCommand(cmds []Command) Command {
	return Command{
		Triggers:    []string{"help", "?"},
		Usage:       "help",
		Description: "show this help message",
		Action: func(r *Runner, args []string) error {
			for _, cmd := range cmds {
				io.WriteString(r.Writer, cmd.Usage)
				if len(cmd.Usage) < 8 {
					r.Writer.Write(tab)
				} else {
					r.Writer.Write(lineBreak)
					r.Writer.Write(tab)
				}
				io.WriteString(r.Writer, cmd.Description)
				io.WriteString(r.Writer, " (triggers: ")
				for i, trigger := range cmd.Triggers {
					if i > 0 {
						r.Writer.Write(comma)
					}
					io.WriteString(r.Writer, trigger)
				}
				io.WriteString(r.Writer, ")\n")
			}
			return nil
		},
	}
}

func (r *Runner) WriteTrans() {
	r.Writer.Write(transPrefix)

	if r.Env.IsCompleted(r.State) {
		io.WriteString(r.Writer, "\tthis is a completed state. no transitions\n")
		return
	}

	trans := r.Env.Transitions(r.State)
	if trans.Len() == 0 {
		io.WriteString(r.Writer, "\there are no transitions. deadlock!\n")
		return
	}

	for i, trans := range trans.Iter() {
		r.Writer.Write(tab)
		io.WriteString(r.Writer, strconv.Itoa(i))
		r.Writer.Write(colon)
		r.Writer.Write(lineBreak)
		r.WriteAllocation(trans.Allocation)
		r.Writer.Write(lineBreak)
	}
}

var TransCommand = Command{
	Triggers:    []string{"trans", "t"},
	Usage:       "trans",
	Description: "show the transitions",
	Action: func(r *Runner, args []string) error {
		r.WriteTrans()
		return nil
	},
}

var ShowStateCommand = Command{
	Triggers:    []string{"show", "s"},
	Usage:       "show",
	Description: "show the state",
	Action: func(r *Runner, args []string) error {
		r.WriteState(r.State)
		return nil
	},
}

var NextStateCommand = Command{
	Triggers:    []string{"next", "n"},
	Usage:       "next <index>",
	Description: "choose next state",
	Action: func(r *Runner, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("fsmrun.NextStateCommand: one argument is required")
		}

		idx, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("fsmrun.NextStateCommand: invalid argument: %w", err)
		}
		if idx < 0 {
			return fmt.Errorf("fsmrun.NextStateCommand: index must be non-negative: %d", idx)
		}

		if err := r.NextState(idx); err != nil {
			return fmt.Errorf("fsmrun.NextStateCommand: %w", err)
		}

		r.WriteCurrent()
		r.WriteTrans()
		return nil
	},
}

var GreedyNextStateCommand = Command{
	Triggers:    []string{"greedy-next", "g"},
	Usage:       "greedy-next [<count>]",
	Description: "choose next state greedily",
	Action: func(r *Runner, args []string) error {

		count := 0
		if len(args) < 1 {
			count = 1
		} else if len(args) > 1 {
			return fmt.Errorf("fsmrun.GreedyNextStateCommand: too many arguments: %v", args)
		} else {
			var err error
			count, err = strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("fsmrun.GreedyNextStateCommand: invalid argument: %w", err)
			}
		}

		for i := 0; i < count; i++ {
			trans := r.Env.Transitions(r.State)

			best, ok := trans.At(0)
			if !ok {
				return fmt.Errorf("fsmrun.GreedyNextStateCommand: no transitions")
			}
			for i, t := range trans.Iter() {
				if i == 0 {
					continue
				}
				if fsm.CompareAllocationByTotalConsumedVolume(t.Allocation, best.Allocation) < 0 {
					best = t
				}
			}
			r.CheckoutState(best)

			r.WriteCurrent()
			r.WriteTrans()
		}
		return nil
	},
}

var EnvCommand = Command{
	Triggers:    []string{"env", "e"},
	Usage:       "env",
	Description: "show the environment",
	Action: func(r *Runner, args []string) error {
		r.WriteEnv()
		return nil
	},
}

var HistoryCommand = Command{
	Triggers:    []string{"history", "h"},
	Usage:       "history [<index>]",
	Description: "show the history or rollback to a specific state. if <index> is negative, rollback to the state at the given index from the end of the history",
	Action: func(r *Runner, args []string) error {
		if len(args) == 0 {
			for i, entry := range r.History {
				io.WriteString(r.Writer, strconv.Itoa(i))
				r.Writer.Write(colon)
				r.Writer.Write(lineBreak)
				r.WriteState(entry.NextState)
				r.Writer.Write(lineBreak)
			}
			return nil
		}

		if len(args) > 1 {
			return fmt.Errorf("fsmrun.HistoryCommand: too many arguments: %v", args)
		}

		idx, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("fsmrun.HistoryCommand: invalid argument: %w", err)
		}
		if idx < -len(r.History) || len(r.History) <= idx {
			return fmt.Errorf("fsmrun.HistoryCommand: index out of range: %d", idx)
		}
		if idx < 0 {
			idx = len(r.History) + idx
		}

		tr := r.History[idx]
		r.History = r.History[:idx]
		r.CheckoutState(tr)

		r.WriteCurrent()
		r.WriteTrans()
		return nil
	},
}

var PrintPlanCommand = Command{
	Triggers:    []string{"plan", "p"},
	Usage:       "plan <path>",
	Description: "Print a plan",
	Action: func(r *Runner, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("fsmrun.PrintPlanCommand: one argument is required")
		}
		if len(args) > 1 {
			return fmt.Errorf("fsmrun.PrintPlanCommand: too many arguments: %v", args)
		}

		path := args[0]
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("fsmrun.PrintPlanCommand: %w", err)
		}
		defer f.Close()

		plan := fsm.NewEmptyPlan(r.Env.InitialState())
		for _, tr := range r.History {
			plan.Add(&fsm.Trans{
				Allocation: tr.Allocation,
				NextState:  tr.NextState,
			})
		}

		e := json.NewEncoder(f)
		e.SetEscapeHTML(false)
		e.SetIndent("", "  ")
		if err := e.Encode(plan); err != nil {
			return fmt.Errorf("fsmrun.PrintPlanCommand: %w", err)
		}
		io.WriteString(r.Writer, "plan write to: ")
		io.WriteString(r.Writer, path)
		r.Writer.Write(lineBreak)
		return nil
	},
}

var QuitCommand = Command{
	Triggers:    []string{"quit", "q", "exit"},
	Usage:       "quit",
	Description: "quit the program",
	Action: func(r *Runner, args []string) error {
		return ErrQuit
	},
}

var AssumeRemainedVolumeCommand = Command{
	Triggers:    []string{"assume-remained-volume"},
	Usage:       "assume-remained-volume <atomic-process> <volume>",
	Description: "assume a remained volume of an atomic process",
	Action: func(r *Runner, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("fsmrun.AssumeRemainedVolumeCommand: two argument is required")
		}
		if len(args) > 2 {
			return fmt.Errorf("fsmrun.AssumeRemainedVolumeCommand: too many arguments: %v", args)
		}

		unsafeAP := args[0]
		found := false
		for _, ap := range r.Env.PFD.AtomicProcesses.Iter() {
			if string(ap) == unsafeAP {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("fsmrun.AssumeRemainedVolumeCommand: atomic process not found: %q", unsafeAP)
		}
		ap := pfd.AtomicProcessID(unsafeAP)

		vol, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("fsmrun.AssumeRemainedVolumeCommand: invalid argument: %w", err)
		}
		r.State.RemainedVolumeMap[ap] = max(fsm.Volume(vol), fsm.MinimumVolume)

		r.WriteCurrent()
		r.WriteTrans()
		return nil
	},
}

var AssumeDeliverableUpdatedCommand = Command{
	Triggers:    []string{"assume-deliverable-updated"},
	Usage:       "assume-deliverable-updated <atomic-deliverable>",
	Description: "assume deliverable is updated",
	Action: func(r *Runner, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("fsmrun.AssumeDeliverableUpdatedCommand: two argument is required")
		}
		if len(args) > 1 {
			return fmt.Errorf("fsmrun.AssumeDeliverableUpdatedCommand: too many arguments: %v", args)
		}

		unsafeD := args[0]
		found := false
		for _, d := range r.Env.PFD.AtomicDeliverables.Iter() {
			if string(d) == unsafeD {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("fsmrun.AssumeDeliverableUpdatedCommand: deliverable not found: %q", unsafeD)
		}

		d := pfd.AtomicDeliverableID(unsafeD)
		for _, ap := range r.Env.PFD.EitherFeedbackOrNotDestinationAtomicProcesses(d).Iter() {
			r.State.UpdatedDeliverablesNotHandled[ap].Add(pfd.AtomicDeliverableID.Compare, d)
		}
		return nil
	},
}

var PrintPreconditionEvalResultCommand = Command{
	Triggers:    []string{"pp", "print-precondition-eval-result"},
	Usage:       "print-precondition-eval-result <atomic-process>...",
	Description: "print the precondition eval result",
	Action: func(r *Runner, args []string) error {
		var aps *sets.Set[pfd.AtomicProcessID]
		if len(args) == 0 {
			aps = r.Env.PFD.AtomicProcesses
		} else {
			for _, arg := range args {
				found := false
				for _, ap := range r.Env.PFD.AtomicProcesses.Iter() {
					if string(ap) == arg {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("fsmrun.PrintPreconditionEvalResultCommand: atomic process not found: %q", arg)
				}
				aps.Add(pfd.AtomicProcessID.Compare, pfd.AtomicProcessID(arg))
			}
		}
		for _, ap := range aps.Iter() {
			r.Writer.Write(doubleTab)
			io.WriteString(r.Writer, string(ap))
			r.Writer.Write(tabArrow)
			r.Env.PreconditionMap[ap].Eval(r.Env, r.State.RemainedVolumeMap, r.State.RevisionMap, r.State.AllocationShouldContinue, r.State.UpdatedDeliverablesNotHandled).Write(r.Writer)
			r.Writer.Write(lineBreak)
		}
		return nil
	},
}

var DefaultCommandsExceptHelp = []Command{
	QuitCommand,
	NextStateCommand,
	GreedyNextStateCommand,
	ShowStateCommand,
	TransCommand,
	EnvCommand,
	HistoryCommand,
	PrintPlanCommand,
	PrintPreconditionEvalResultCommand,
	AssumeRemainedVolumeCommand,
	AssumeDeliverableUpdatedCommand,
}

var DefaultCommands = append(slices.Clone(DefaultCommandsExceptHelp), NewHelpCommand(DefaultCommandsExceptHelp))

func NewRunner(reader io.Reader, writer io.Writer, env *fsm.Env, commands []Command) *Runner {
	initState := env.InitialState()
	return &Runner{Reader: reader, Writer: writer, Env: env, State: initState, History: []*fsm.Trans{}, Commands: commands}
}

func (r *Runner) NextState(idx int) error {
	trans := r.Env.Transitions(r.State)
	if idx >= trans.Len() {
		return fmt.Errorf("fsmrun.NextStateCommand: index out of range: %d", idx)
	}
	tr, ok := trans.At(idx)
	if !ok {
		return fmt.Errorf("fsmrun.NextStateCommand: index out of range: %d", idx)
	}
	r.CheckoutState(tr)
	return nil
}

func (r *Runner) CheckoutState(tr *fsm.Trans) {
	r.State = tr.NextState
	r.History = append(r.History, tr)
}

func (r *Runner) ApplyPlan(plan *fsm.Plan) error {
	for _, tr := range plan.Transitions {
		r.CheckoutState(tr)
	}
	return nil
}

func (r *Runner) Run(plan *fsm.Plan) error {
	r.WriteEnv()

	if plan != nil {
		if err := r.ApplyPlan(plan); err != nil {
			return fmt.Errorf("fsmrun.Run: %w", err)
		}
	}

	r.WriteCurrent()
	r.WriteTrans()
	r.WritePrompt()

	scanner := bufio.NewScanner(r.Reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			r.WritePrompt()
			continue
		}
		parts := strings.Split(line, " ")
		command := parts[0]
		args := parts[1:]
		handled := false
		for _, cmd := range r.Commands {
			if slices.Contains(cmd.Triggers, command) {
				handled = true
				if err := cmd.Action(r, args); err != nil {
					if errors.Is(err, ErrQuit) {
						return nil
					}
					r.Writer.Write(errorPrefix)
					io.WriteString(r.Writer, err.Error())
					r.Writer.Write(lineBreak)
				}
				break
			}
		}
		if !handled {
			fmt.Fprintf(r.Writer, "unknown command: %q\n", command)
		}
		r.WritePrompt()
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("fsmrun.Runner.Run: %w", err)
	}
	return nil
}

func (r *Runner) WritePrompt() {
	r.Writer.Write(promptPrefix)
}

func (r *Runner) WriteEnv() {
	r.Writer.Write(envPrefix)

	r.Writer.Write(availableResourcesPrefix)
	r.Writer.Write(doubleTab)
	for i, resource := range r.Env.AvailableResources.Iter() {
		if i > 0 {
			r.Writer.Write(comma)
		}
		io.WriteString(r.Writer, string(resource))
	}
	r.Writer.Write(lineBreak)

	r.Writer.Write(initialVolumePrefix)
	for _, ap := range r.Env.PFD.AtomicProcesses.Iter() {
		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(ap))
		r.Writer.Write(tabArrow)
		io.WriteString(r.Writer, r.Env.InitialVolumeFunc(ap).String())
		r.Writer.Write(lineBreak)
	}

	r.Writer.Write(reworkVolumePrefix)
	for _, ap := range r.Env.PFD.AtomicProcesses.Iter() {
		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(ap))
		r.Writer.Write(tabArrow)
		for i := 0; i < 100; i++ {
			if i > 0 {
				r.Writer.Write(arrow)
			}
			vol := r.Env.ReworkVolumeFunc(ap, i)
			io.WriteString(r.Writer, vol.String())
			if vol == fsm.MinimumVolume {
				break
			}
		}
		r.Writer.Write(lineBreak)
	}

	r.Writer.Write(neededResourceSetsPrefix)
	for _, ap := range r.Env.PFD.AtomicProcesses.Iter() {
		for _, entry := range r.Env.NeededResourceSetsFunc(ap).Iter() {
			r.Writer.Write(doubleTab)
			io.WriteString(r.Writer, string(ap))
			r.Writer.Write(tabArrow)
			for i, resource := range entry.Resources.Iter() {
				if i > 0 {
					r.Writer.Write(comma)
				}
				io.WriteString(r.Writer, string(resource))
			}
			r.Writer.Write(comma)
			io.WriteString(r.Writer, strconv.Itoa(int(entry.ConsumedVolume)))
			r.Writer.Write(semicolon)
			r.Writer.Write(lineBreak)
		}
	}

	r.Writer.Write(deliverableAvailableTimePrefix)
	for _, d := range r.Env.PFD.InitialDeliverables().Iter() {
		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(d))
		r.Writer.Write(tabArrow)
		io.WriteString(r.Writer, strconv.Itoa(int(r.Env.DeliverableAvailableTimeFunc(d))))
		r.Writer.Write(lineBreak)
	}

	r.Writer.Write(preconditionPrefix)
	for _, p := range sortedKVFromMap(r.Env.PreconditionMap, pfd.AtomicProcessID.Compare) {
		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(p.First))
		r.Writer.Write(tabArrow)
		p.Second.Compile(r.Env.PFD, r.Env.Logger).Write(r.Writer)
		r.Writer.Write(lineBreak)
	}
}

func (r *Runner) WriteCurrent() {
	r.Writer.Write(currentPrefix)
	r.WriteState(r.State)
}

func (r *Runner) WriteState(state fsm.State) error {
	r.Writer.Write(timePrefix)
	io.WriteString(r.Writer, strconv.FormatFloat(float64(state.Time), 'f', -1, 64))
	r.Writer.Write(lineBreak)

	r.Writer.Write(remainedVolumePrefix)
	for _, pair := range sortedKVFromMap(state.RemainedVolumeMap, pfd.AtomicProcessID.Compare) {
		ap := pair.First
		remainedVolume := pair.Second

		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(ap))
		r.Writer.Write(tabArrow)
		io.WriteString(r.Writer, remainedVolume.String())
		r.Writer.Write(lineBreak)
	}

	r.Writer.Write(numOfExecutionPrefix)
	for _, pair := range sortedKVFromMap(state.NumOfCompleteMap, pfd.AtomicProcessID.Compare) {
		ap := pair.First
		numOfExecution := pair.Second

		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(ap))
		r.Writer.Write(tabArrow)
		io.WriteString(r.Writer, strconv.Itoa(numOfExecution))
		r.Writer.Write(lineBreak)
	}

	r.Writer.Write(revisionPrefix)
	for _, pair := range sortedKVFromMap(state.RevisionMap, pfd.AtomicDeliverableID.Compare) {
		d := pair.First
		revision := pair.Second

		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(d))
		r.Writer.Write(tabArrow)
		io.WriteString(r.Writer, strconv.Itoa(revision))
		r.Writer.Write(lineBreak)
	}

	r.Writer.Write(allocationPrefix)
	r.WriteAllocation(state.AllocationShouldContinue)

	r.Writer.Write(allocatabilityPrefix)
	for _, pair := range sortedKVFromMap(r.Env.AllocatabilityInfoMap(state), pfd.AtomicProcessID.Compare) {
		ap := pair.First
		allocatabilityInfo := pair.Second
		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(ap))
		r.Writer.Write(tabArrow)
		allocatabilityInfo.Write(r.Writer)
		r.Writer.Write(lineBreak)
	}

	return nil
}

func (r *Runner) WriteAllocation(allocation fsm.Allocation) error {
	if len(allocation) == 0 {
		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, "none")
		r.Writer.Write(lineBreak)
		return nil
	}

	for _, pair := range sortedKVFromMap(allocation, pfd.AtomicProcessID.Compare) {
		ap := pair.First
		allocationElement := pair.Second

		r.Writer.Write(doubleTab)
		io.WriteString(r.Writer, string(ap))
		r.Writer.Write(tabArrow)
		for i, resource := range allocationElement.Resources.Iter() {
			if i > 0 {
				r.Writer.Write(comma)
			}
			io.WriteString(r.Writer, string(resource))
		}
		r.Writer.Write(comma)
		io.WriteString(r.Writer, strconv.Itoa(int(allocationElement.ConsumedVolume)))
		r.Writer.Write(semicolon)
		r.Writer.Write(lineBreak)
	}
	return nil
}

func sortedKVFromMap[K comparable, V any](m map[K]V, compare func(K, K) int) []pairs.Pair[K, V] {
	xs := make([]pairs.Pair[K, V], 0, len(m))
	for k := range m {
		xs = append(xs, pairs.Pair[K, V]{First: k, Second: m[k]})
	}
	slices.SortFunc(xs, func(a, b pairs.Pair[K, V]) int {
		return compare(a.First, b.First)
	})
	return xs
}

var (
	envPrefix                      = []byte("env:\n")
	currentPrefix                  = []byte("current:\n")
	transPrefix                    = []byte("trans:\n")
	timePrefix                     = []byte("\ttime: ")
	remainedVolumePrefix           = []byte("\tremained volume:\n")
	revisionPrefix                 = []byte("\trevision:\n")
	numOfExecutionPrefix           = []byte("\tnum of complete:\n")
	allocationPrefix               = []byte("\tallocation should continue:\n")
	availableResourcesPrefix       = []byte("\tavailable resources:\n")
	initialVolumePrefix            = []byte("\tinitial volume:\n")
	reworkVolumePrefix             = []byte("\trework volume:\n")
	neededResourceSetsPrefix       = []byte("\tneeded resource sets:\n")
	deliverableAvailableTimePrefix = []byte("\tdeliverable available time:\n")
	allocatabilityPrefix           = []byte("\tallocatability:\n")
	preconditionPrefix             = []byte("\tprecondition:\n")
	promptPrefix                   = []byte("> ")
	errorPrefix                    = []byte("error: ")
	lineBreak                      = []byte("\n")
	arrow                          = []byte(" -> ")
	tabArrow                       = []byte("\t-> ")
	tab                            = []byte("\t")
	doubleTab                      = []byte("\t\t")
	comma                          = []byte(", ")
	semicolon                      = []byte(";")
	colon                          = []byte(": ")
)
