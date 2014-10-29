package checks

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
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

func (pc *ProcessCheck) excludeProcesses(list []process) []process {
	processes := make([]process, 0, len(list))
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
	var trapProcUsers []string
	if "" != pc.user {
		trapProcUsers = strings.Split(pc.user, ",")
	} else {
		trapProcUsers = make([]string, 0)
	}

	fmt.Printf("---- list is %d items long\n", len(list))

	for _, proc := range list {
		//fmt.Print(" -> ")
		if !pc.matchSelf && pid == proc.pid {
			fmt.Println("skipping matchParent")
			continue
		}

		if !pc.matchParent && ppid == proc.pid {
			fmt.Println("skipping matchParent")
			continue
		}

		if pc.commandPattern != "" && !regex.MatchString(proc.command) {
			fmt.Println("skipping commandPattern")
			continue
		}

		if pc.vszErrorOver >= 0 && proc.vsz > pc.vszErrorOver {
			fmt.Println("skipping vszErrorOver")
			continue
		}
		if pc.rssErrorOver >= 0 && proc.rss > pc.rssErrorOver {
			fmt.Println("skipping rssErrorOver")
			continue
		}
		if pc.pcpuErrorOver >= 0.0 && proc.pcpu > pc.pcpuErrorOver {
			fmt.Println("skipping pcpuErrorOver")
			continue
		}
		if pc.threadsErrOver >= 0 && proc.threads > pc.threadsErrOver {
			fmt.Println("skipping threadsErrOver")
			continue
		}

		if pc.processessYoungerThan >= 0 && proc.processTime >= pc.processessYoungerThan {
			fmt.Println("skipping processessYoungerThan")
			continue
		}

		if pc.processessOlderThan >= 0 && proc.processTime <= pc.processessOlderThan {
			fmt.Println("skipping processessOlderThan")
			continue
		}

		if pc.cpuTimeLessThan >= 0 && proc.cpuTime >= pc.cpuTimeLessThan {
			fmt.Println("skipping cpuTimeLessThan")
			continue
		}
		if pc.cpuTimeMoreThan >= 0 && proc.cpuTime <= pc.cpuTimeMoreThan {
			fmt.Println("skipping cpuTimeMoreThan")
			continue
		}

		// match process states to ones we want
		skip := false
		for _, state := range trapProcStats {
			if proc.state == state {
				fmt.Println("skipping: matching state")
				skip = true
				break
			}
		}
		for _, user := range trapProcUsers {
			if proc.user == user {
				fmt.Println("skipping: matching user")
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

func (pc *ProcessCheck) gatherProcesses() []process {
	var processes []process
	var err error
	var status map[string]string

	paths, err := filepath.Glob("/proc/[0-9]*/status")
	if nil != err {
		log.Printf("Unable to get a list of processes. %s", err)
		return processes
	}

	for _, path := range paths {
		proc := new(process)
		status = fileToMap(path)
		proc.pid = atoi(status["Pid"])
		proc.ppid = atoi(status["PPid"])

		proc.uid = status["Uid"]
		u, err := user.LookupId(status["Uid"])
		if err == nil {
			proc.user = u.Username
		}
		proc.state = status["State"]

		proc.threads = atoi(status["Threads"])
		proc.rss = atoi(status["VmRSS"])
		proc.vsz = atoi(status["VmSize"])

		cmdline, _ := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", proc.pid))
		proc.command = string(cmdline)
		//proc.pcpu

		stat := getProcStat(fmt.Sprintf("/proc/%d/stat", proc.pid))
		proc.cpuTime = stat["utime"]
		proc.processTime = stat["cutime"]

		processes = append(processes, *proc)
	}

	return processes
}

func atoi(s string) int {
	i, _ := strconv.ParseInt(s, 10, 32)
	return int(i)
}

func getProcStat(file string) map[string]int {
	ret := make(map[string]int)
	var pid, ppid, pgrp, session, tty_nr, tpgid, flags, minflt, cminflt, majflt, cmajflt, utime, stime, cutime, cstime, priority int
	var exe, status string
	content, _ := ioutil.ReadFile(file)

	fmt.Sscanf(string(content), "%d %s %c %d %d %d %d %d %d %d %d %d %d %d %d %d %d", &pid, &exe, &status, &ppid, &pgrp, &session, &tty_nr, &tpgid, &flags, &minflt, &cminflt, &majflt, &cmajflt, &utime, &stime, &cutime, &cstime, &priority)

	ret["pid"] = pid
	ret["ppid"] = ppid
	ret["pgrp"] = pgrp
	ret["session"] = session
	ret["tty_nr"] = tty_nr
	ret["tpgid"] = tpgid
	ret["flags"] = flags
	ret["minflt"] = minflt
	ret["cminflt"] = cminflt
	ret["majflt"] = majflt
	ret["cmajflt"] = cmajflt
	ret["utime"] = utime
	ret["stime"] = stime
	ret["cutime"] = cutime
	ret["cstime"] = cstime
	ret["priority"] = priority
	return ret
}

// helper function to turn a file into a kv map
func fileToMap(file string) map[string]string {
	var content []byte
	var lines, fields []string
	var err error
	ret := make(map[string]string)
	content, err = ioutil.ReadFile(file)
	lines = strings.Split(string(content), "\n")

	r, err := regexp.Compile("[:\\s\\t]+")
	if nil != err {
		log.Printf("Cannot compile regex: %s", err)
	}

	for _, line := range lines {
		fields = r.Split(line, 4)
		if len(fields) < 2 {
			continue
		}
		ret[fields[0]] = fields[1]
	}
	return ret
}
