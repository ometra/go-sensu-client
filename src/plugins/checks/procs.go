package checks

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"plugins"
	"regexp"
	"strconv"
	"strings"
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
	metric                                     string
	matchSelf, matchParent                     bool
	commandPattern, filePid                    string
	filePidActual                              int64
	vszErrorOver, rssErrorOver, threadsErrOver int
	pcpuErrorOver                              float64
	processState, user                         string
	processessOlderThan, processessYoungerThan int
	cpuTimeMoreThan, cpuTimeLessThan           int

	ShowHelp    bool
	checkStatus plugins.Status
}

type process struct {
	pid, ppid int // process id and parent process id
	command   string

	vsz, rss, threads    int
	pcpu                 float64
	cpuTime, processTime int // in seconds

	state, uid, user string
}

func (pc *ProcessCheck) Init(config plugins.PluginConfig) (string, error) {
	pc.flags = flag.NewFlagSet("process-check", flag.ContinueOnError)

	pc.addFlag("w", "-warn-over", "Trigger a warning if over a number", &pc.warnOver, -1)
	pc.addFlag("c", "-critical-over", "Trigger a critical if over a number", &pc.criticalOver, -1)
	pc.addFlag("W", "-warn-under", "Trigger a warning if under a number", &pc.warnUnder, 1)
	pc.addFlag("C", "-critical-under", "Trigger a critical if under a number", &pc.criticalUnder, 1)
	pc.addFlag("t", "-metric", "Count and return one of [vsz,rss,threads,pcpu,cpuTime,processTime]. obeys warn/crit thresholds", &pc.metric, "")
	pc.addFlag("m", "-match-self", "Include this script in the list", &pc.matchSelf, false)
	pc.addFlag("M", "-match-parent", "ignored for compatability (no ruby parent)", &pc.matchParent, false)
	pc.addFlag("p", "-pattern", "Include commands matching this regexp pattern", &pc.commandPattern, "")
	pc.addFlag("f", "-file-pid", "Check against a specific PID", &pc.filePid, "")
	pc.addFlag("z", "-virtual-memory-size", "Ignore processes with a Virtual Memory size is bigger than this", &pc.vszErrorOver, -1)
	pc.addFlag("r", "-resident-set-size", "Ignore processes with a Resident Set size is bigger than this", &pc.rssErrorOver, -1)
	pc.addFlag("P", "-proportional-set-size", "Ignore processes with a Proportional Set Size is bigger than this", &pc.pcpuErrorOver, -1.0)
	pc.addFlag("T", "-thread-count", "Ignore processes with a Thread Count is bigger than this", &pc.threadsErrOver, -1)
	pc.addFlag("s", "-state", "Ignore processes with a specific state, example: Z for zombie. Comma seperated list", &pc.processState, "")
	pc.addFlag("u", "-user", "Ignore processes with a specific user. Comma seperated list", &pc.user, "")
	pc.addFlag("e", "-esec-over", "Match processes that older that this, in SECONDS", &pc.processessOlderThan, -1)
	pc.addFlag("E", "-esec-under", "Match process that are younger than this, in SECONDS", &pc.processessYoungerThan, -1)
	pc.addFlag("i", "-cpu-over", "Match processes cpu time that is older than this, in SECONDS", &pc.cpuTimeMoreThan, -1)
	pc.addFlag("I", "-cpu-under", "Match processes cpu time that is younger than this, in SECONDS", &pc.cpuTimeLessThan, -1)

	pc.addFlag("h", "-help", "Show help", &pc.ShowHelp, false)

	err := pc.flags.Parse(config.Args[1:])
	if nil != err {
		log.Printf("Failed to parse process check command line: %s", err)
	}

	return "process-check", nil
}

func (pc *ProcessCheck) addFlag(short, long, description string, target interface{}, defaultValue interface{}) {
	switch t := target.(type) {
	case *int:
		pc.flags.IntVar(t, short, defaultValue.(int), description)
		pc.flags.IntVar(t, long, defaultValue.(int), description)
	case *float64:
		pc.flags.Float64Var(t, short, defaultValue.(float64), description)
		pc.flags.Float64Var(t, long, defaultValue.(float64), description)
	case *bool:
		pc.flags.BoolVar(t, short, defaultValue.(bool), description)
		pc.flags.BoolVar(t, long, defaultValue.(bool), description)
	case *string:
		pc.flags.StringVar(t, short, defaultValue.(string), description)
		pc.flags.StringVar(t, long, defaultValue.(string), description)
	}
}

// sets the check status
func (pc *ProcessCheck) SetCheckStatus(status plugins.Status) {
	pc.checkStatus = status
}

// handles the gathering of data
func (pc *ProcessCheck) Gather(r *plugins.Result) error {
	pc.SetCheckStatus(plugins.UNKNOWN)
	r.SetStatus(plugins.UNKNOWN)
	err := pc.createPayload(r)
	return err
}

// shows the usage
func (pc *ProcessCheck) Usage() {
	pc.flags.PrintDefaults()
}

// gets a formatted status
func (pc *ProcessCheck) GetStatus() string {
	return "CheckProcs " + pc.checkStatus.ToString()
}

// excludes processes from the list
func (pc *ProcessCheck) excludeProcesses(list []process) []process {
	processes := make([]process, 0, len(list))
	var regex *regexp.Regexp
	var err error

	pid := os.Getpid()
	//ppid := os.Getppid()

	if pc.commandPattern != "" {
		regex, err = regexp.Compile(pc.commandPattern)
		if nil != err {
			log.Printf("Invalid Regexp Pattern: (%s) %s\n", pc.commandPattern, err)
			os.Exit(1)
		}
	}

	trapProcStats := strings.Split(pc.processState, ",")
	var trapProcUsers []string
	if "" != pc.user {
		trapProcUsers = strings.Split(pc.user, ",")
	} else {
		trapProcUsers = make([]string, 0)
	}

	//fmt.Printf("---- list is %d items long\n", len(list))

	for _, proc := range list {
		//fmt.Print(" -> ")
		if !pc.matchSelf && pid == proc.pid {
			//fmt.Println("skipping matchParent")
			continue
		}

		// we do not need to ignore the parent process (the original uses this
		// to ignore the Ruby proc that is running the script
		//if !pc.matchParent && ppid == proc.pid {
		//	//fmt.Println("skipping matchParent")
		//	continue
		//}

		if pc.filePid != "" && proc.pid != int(pc.filePidActual) {
			continue
		}

		if pc.commandPattern != "" && !regex.MatchString(proc.command) {
			//fmt.Println("skipping commandPattern")
			continue
		}

		if pc.vszErrorOver >= 0 && proc.vsz > pc.vszErrorOver {
			//fmt.Println("skipping vszErrorOver")
			continue
		}
		if pc.rssErrorOver >= 0 && proc.rss > pc.rssErrorOver {
			//fmt.Println("skipping rssErrorOver")
			continue
		}
		if pc.pcpuErrorOver >= 0.0 && proc.pcpu > pc.pcpuErrorOver {
			//fmt.Println("skipping pcpuErrorOver")
			continue
		}
		if pc.threadsErrOver >= 0 && proc.threads > pc.threadsErrOver {
			//fmt.Println("skipping threadsErrOver")
			continue
		}

		if pc.processessYoungerThan >= 0 && proc.processTime >= pc.processessYoungerThan {
			//fmt.Println("skipping processessYoungerThan")
			continue
		}

		if pc.processessOlderThan >= 0 && proc.processTime <= pc.processessOlderThan {
			//fmt.Println("skipping processessOlderThan")
			continue
		}

		if pc.cpuTimeLessThan >= 0 && proc.cpuTime >= pc.cpuTimeLessThan {
			//fmt.Println("skipping cpuTimeLessThan")
			continue
		}
		if pc.cpuTimeMoreThan >= 0 && proc.cpuTime <= pc.cpuTimeMoreThan {
			//fmt.Println("skipping cpuTimeMoreThan")
			continue
		}

		// match process states to ones we want
		skip := false
		for _, state := range trapProcStats {
			if proc.state == state {
				//fmt.Println("skipping: matching state")
				skip = true
				break
			}
		}
		for _, user := range trapProcUsers {
			if proc.user == user {
				//fmt.Println("skipping: matching user")
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		//fmt.Println("Success - found a proc to count")
		processes = append(processes, proc)
	}

	return processes
}

// make our return payload
func (pc *ProcessCheck) createPayload(r *plugins.Result) error {

	if pc.filePid != "" {
		content, err := ioutil.ReadFile(pc.filePid)
		if nil != err {
			r.Add("Could not read pid file " + pc.filePid)
			return err
		}
		s := strings.Trim(string(content), "\n ")
		pc.filePidActual, _ = strconv.ParseInt(s, 10, 32)
	}

	process_list := pc.excludeProcesses(pc.gatherProcesses())

	r.Add(fmt.Sprintf("Found %d matching processes", len(process_list)))

	if pc.commandPattern != "" {
		r.Add(fmt.Sprintf("; cmd /%s/", pc.commandPattern))
	}

	if pc.processState != "" {
		r.Add(fmt.Sprintf("; state %s", pc.processState))
	}

	if pc.user != "" {
		r.Add(fmt.Sprintf("; user %s", pc.user))
	}

	if pc.vszErrorOver > 0 {
		r.Add(fmt.Sprintf("; vsz < %d", pc.vszErrorOver))
	}

	if pc.rssErrorOver > 0 {
		r.Add(fmt.Sprintf("; rss < %d", pc.rssErrorOver))
	}

	if pc.pcpuErrorOver > 0 {
		r.Add(fmt.Sprintf("; pcpu < %d", pc.pcpuErrorOver))
	}

	if pc.threadsErrOver > 0 {
		r.Add(fmt.Sprintf("; thcount < %d", pc.threadsErrOver))
	}

	if pc.processessYoungerThan > 0 {
		r.Add(fmt.Sprintf("; esec < %d", pc.processessYoungerThan))
	}

	if pc.processessOlderThan > 0 {
		r.Add(fmt.Sprintf("; esec > %d", pc.processessOlderThan))
	}

	if pc.cpuTimeLessThan > 0 {
		r.Add(fmt.Sprintf("; csec < %d", pc.cpuTimeLessThan))
	}

	if pc.cpuTimeMoreThan > 0 {
		r.Add(fmt.Sprintf("; csec > %d", pc.cpuTimeMoreThan))
	}

	if pc.filePid != "" {
		r.Add(fmt.Sprintf("; pid %s", pc.filePid))
	}

	var count int
	if pc.metric != "" {
		count = 0
		for _, proc := range process_list {

			switch pc.metric {
			case "vsz":
				count += proc.vsz
			case "rss":
				count += proc.rss
			case "threads", "thcount":
				count += proc.threads
			case "pcpu":
				count += int(proc.pcpu)
			case "time":
				count += proc.cpuTime
			case "etime":
				count += proc.processTime
			}
		}

		r.Add(fmt.Sprintf("; %s == %d", pc.metric, count))
	} else {
		count = len(process_list)
	}

	if pc.criticalUnder > 0 && count < pc.criticalUnder {
		// return a critical response
		pc.SetCheckStatus(plugins.CRITICAL)
	} else if pc.criticalOver >= 0 && count > pc.criticalOver {
		// return a critical response
		pc.SetCheckStatus(plugins.CRITICAL)
	} else if pc.warnUnder > 0 && count < pc.warnUnder {
		// return a warning response
		pc.SetCheckStatus(plugins.WARNING)
	} else if pc.warnOver >= 0 && count > pc.warnOver {
		// return a warning response
		pc.SetCheckStatus(plugins.WARNING)
	} else {
		// everything is ok
		pc.SetCheckStatus(plugins.OK)
	}

	return nil
}
