package checks

import (
	"os"
	"testing"
)

func TestExcludeProcs(t *testing.T) {
	var list, testList []process
	var testLen int

	proc := process{
		pid:         os.Getpid(),
		ppid:        os.Getppid(),
		command:     "moo --is --you",
		vsz:         33912,
		rss:         4164,
		pcpu:        0.0,
		threads:     65,
		cpuTime:     99999,
		processTime: 100,
		state:       "S",
		user:        "root",
	}

	matchlist := append(list, proc)

	pc := new(ProcessCheck)

	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 0 != testLen {
		t.Errorf("Failed to exclude myself from the list of processes, expect 0 results, got %d.", testLen)
	}

	pc.matchSelf = true
	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 1 != testLen {
		t.Errorf("Failed to include myself from the list of processes, expect 0 results, got %d.", testLen)
	}
}
