package checks

import (
	"flag"
	"log"
)

// Checks to see if a process is running
//
// DESCRIPTION
//  This plugin checks to see if a given process is running
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type ProcessCheck struct {
	flags                                      *flag.FlagSet
	warnOver, criticalOver                     int
	warnUnder, criticalUnder                   int
	metric                                     int
	matchSelf, matchParent                     bool
	commandPattern, filePid                    string
	vszErrorOver, rssErrorOver                 int
	pcpuErrorOver, threadsErrOver              int
	processState, user                         string
	processessOlderThan, processessYoungerThan int
	cpuTimeMoreThan, cpuTimeLessThan           int
}

func (pc *ProcessCheck) Init(config CheckConfigType) (string, error) {
	pc.flags = flag.NewFlagSet("process-check", flag.ContinueOnError)

	pc.addFlag("w", "-warn-over", "Trigger a warning if over a number", &pc.warnOver, 0)
	pc.addFlag("c", "-critical-over", "Trigger a critical if over a number", &pc.criticalOver, 0)
	pc.addFlag("W", "-warn-under", "Trigger a warning if under a number", &pc.warnUnder, 1)
	pc.addFlag("C", "-critical-under", "Trigger a critical if under a number", &pc.criticalUnder, 1)
	pc.addFlag("t", "-metric", "Trigger a critical if there are METRIC procs", &pc.metric, 0)
	pc.addFlag("m", "-match-self", "Match itself", &pc.matchSelf, false)
	pc.addFlag("M", "-match-parent", "Match parent process it uses ruby {process.ppid}", &pc.matchParent, false)
	pc.addFlag("p", "-pattern", "Match a command against this pattern", &pc.commandPattern, "")
	pc.addFlag("f", "-file-pid", "Check against a specific PID", &pc.filePid, "")
	pc.addFlag("z", "-virtual-memory-size", "Trigger on a Virtual Memory size is bigger than this", &pc.vszErrorOver, 0)
	pc.addFlag("r", "-resident-set-size", "Trigger on a Resident Set size is bigger than this", &pc.rssErrorOver, 0)
	pc.addFlag("P", "-proportional-set-size", "Trigger on a Proportional Set Size is bigger than this", &pc.pcpuErrorOver, 0)
	pc.addFlag("T", "-thread-count", "Trigger on a Thread Count is bigger than this", &pc.threadsErrOver, 0)
	pc.addFlag("s", "-state", "Trigger on a specific state, example: Z for zombie", &pc.processState, "")
	pc.addFlag("u", "-user", "Trigger on a specific user", &pc.user, "")
	pc.addFlag("e", "-esec-over", "Match processes that older that this, in SECONDS", &pc.processessOlderThan, 0)
	pc.addFlag("E", "-esec-under", "Match process that are younger than this, in SECONDS", &pc.processessYoungerThan, 0)
	pc.addFlag("i", "-cpu-over", "Match processes cpu time that is older than this, in SECONDS", &pc.cpuTimeMoreThan, 0)
	pc.addFlag("I", "-cpu-under", "Match processes cpu time that is younger than this, in SECONDS", &pc.cpuTimeLessThan, 0)

	err := pc.flags.Parse(config.Args[1:])
	if nil != err {
		log.Printf("Failed to parse process check command line: %s", err)
	}
	log.Printf("warn over: %d", pc.warnOver)

	return "process-check", nil
}

func (pc *ProcessCheck) addFlag(short, long, description string, target interface{}, defaultValue interface{}) {
	switch t := target.(type) {
	case *int:
		pc.flags.IntVar(t, short, defaultValue.(int), description)
		pc.flags.IntVar(t, long, defaultValue.(int), description)
	case *bool:
		pc.flags.BoolVar(t, short, defaultValue.(bool), description)
		pc.flags.BoolVar(t, long, defaultValue.(bool), description)
	case *string:
		pc.flags.StringVar(t, short, defaultValue.(string), description)
		pc.flags.StringVar(t, long, defaultValue.(string), description)
	}
}

func (pc *ProcessCheck) Gather(r *Result) error {
	output, err := pc.createPayload(r.ShortName(), r.StartTime())
	r.SetOutput(output)
	return err
}

func (pc *ProcessCheck) Usage() {
	pc.flags.PrintDefaults()
}
