package checks

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func (pc *ProcessCheck) createPayload(short_name string, timestamp uint) (string, error) {
	var payload string

	process_list := pc.excludeProcesses(pc.gatherProcesses())

	payload = fmt.Sprintf("Found %d matching processes", len(process_list))

	if pc.commandPattern != "" {
		payload += fmt.Sprintf("; cmd /%s/", pc.commandPattern)
	}

	if pc.processState != "" {
		payload += fmt.Sprintf("; state %s", pc.processState)
	}

	if pc.user != "" {
		payload += fmt.Sprintf("; user %s", pc.user)
	}

	if pc.vszErrorOver > 0 {
		payload += fmt.Sprintf("; vsz < %d", pc.vszErrorOver)
	}

	if pc.rssErrorOver > 0 {
		payload += fmt.Sprintf("; rss < %d", pc.rssErrorOver)
	}

	if pc.pcpuErrorOver > 0 {
		payload += fmt.Sprintf("; pcpu < %d", pc.pcpuErrorOver)
	}

	if pc.threadsErrOver > 0 {
		payload += fmt.Sprintf("; thcount < %d", pc.threadsErrOver)
	}

	if pc.processessYoungerThan > 0 {
		payload += fmt.Sprintf("; esec < %d", pc.processessYoungerThan)
	}

	if pc.processessOlderThan > 0 {
		payload += fmt.Sprintf("; esec > %d", pc.processessOlderThan)
	}

	if pc.cpuTimeLessThan > 0 {
		payload += fmt.Sprintf("; csec < %d", pc.cpuTimeLessThan)
	}

	if pc.cpuTimeMoreThan > 0 {
		payload += fmt.Sprintf("; csec > %d", pc.cpuTimeMoreThan)
	}

	if pc.filePid != "" {
		payload += fmt.Sprintf("; pid %s", pc.filePid)
	}

	var count int
	if pc.metric > 0 {
		count = len(process_list)
		payload += fmt.Sprintf("; %d == %d", pc.metric, count)
	} else {
		count = len(process_list)
	}

	if pc.criticalUnder > 0 && count < pc.criticalUnder {
		// return a critical response
		pc.SetCheckStatus(CRITICAL)
	} else if pc.criticalOver >= 0 && count > pc.criticalOver {
		// return a critical response
		pc.SetCheckStatus(CRITICAL)
	} else if pc.warnUnder > 0 && count < pc.warnUnder {
		// return a warning response
		pc.SetCheckStatus(WARNING)
	} else if pc.warnOver >= 0 && count > pc.warnOver {
		// return a warning response
		pc.SetCheckStatus(WARNING)
	} else {
		// everything is ok
		pc.SetCheckStatus(OK)
	}

	return payload, nil
}

func (pc *ProcessCheck) gatherProcesses() []process {
	var processes []process
	return processes
}

func (pc *ProcessCheck) excludeProcesses(list []process) []process {
	processes := make([]process, 0)
	var regex *regexp.Regexp
	var err error

	pid := os.Getpid()
	ppid := os.Getppid()

	if pc.commandPattern != "" {
		regex, err = regexp.Compile(pc.commandPattern)
		if nil != err {
			log.Printf("Invalid Regexp Pattern: (%s) %s\n", pc.commandPattern, err)
			os.Exit(1)
		}
	}

	trapProcStats := strings.Split(pc.processState, ",")
	trapProcUsers := strings.Split(pc.user, ",")

	for _, proc := range list {
		fmt.Printf("if %v && %d == %d\n", !pc.matchSelf, pid, proc.pid)
		if !pc.matchSelf && pid == proc.pid {
			continue
		}

		if !pc.matchParent && ppid == proc.ppid {
			continue
		}

		if pc.commandPattern != "" && !regex.MatchString(proc.command) {
			continue
		}

		if pc.vszErrorOver > 0 && proc.vsz > pc.vszErrorOver {
			continue
		}
		if pc.rssErrorOver > 0 && proc.rss > pc.rssErrorOver {
			continue
		}
		if pc.pcpuErrorOver > 0 && proc.pcpu > pc.pcpuErrorOver {
			continue
		}
		if pc.threadsErrOver > 0 && proc.threads > pc.threadsErrOver {
			continue
		}

		if pc.processessYoungerThan > 0 && proc.processTime >= pc.processessYoungerThan {
			continue
		}

		if pc.processessOlderThan > 0 && proc.processTime <= pc.processessOlderThan {
			continue
		}

		if pc.cpuTimeLessThan > 0 && proc.cpuTime >= pc.cpuTimeLessThan {
			continue
		}
		if pc.cpuTimeMoreThan > 0 && proc.cpuTime <= pc.cpuTimeMoreThan {
			continue
		}

		// match process states to ones we want
		skip := false
		for _, state := range trapProcStats {
			if proc.state == state {
				skip = true
				break
			}
		}
		for _, user := range trapProcUsers {
			if proc.user == user {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		processes = append(processes, proc)
	}

	return processes
}
