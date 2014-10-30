package checks

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

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
		//proc.pcpu - percent CPU

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
