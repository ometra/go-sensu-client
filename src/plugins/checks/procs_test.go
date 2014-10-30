package checks

import (
	"fmt"
	"os"
	"plugins"
	"testing"
)

var the_check = plugins.PluginConfig{
	Type:       "check",
	Command:    "",
	Handlers:   []string{},
	Standalone: true,
	Interval:   15,
}

func TestExcludeProcs(t *testing.T) {
	var list, testList []process
	var testLen int

	list = append(list, getProc())
	list = append(list, getParentProc())

	matchlist := list

	pc := new(ProcessCheck)

	// make sure we exclude ourselves by default
	the_check.Args = []string{"cmd"}
	pc.Init(the_check)

	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 0 != testLen {
		t.Errorf("Failed to exclude myself from the list of processes, expect 0 results, got %d.", testLen)
	}

	// check to see if we can match ourself in the list
	the_check.Args = []string{"cmd", "-m"}
	pc.Init(the_check)

	fmt.Println("here")
	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 1 != testLen {
		t.Errorf("Failed to include myself in the list of processes, expect 1 results, got %d.", testLen)
	}

	// check to see if we can match our parent in the list
	the_check.Args = []string{"cmd", "-M"}
	pc.Init(the_check)

	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 1 != testLen {
		t.Errorf("Failed to include my parent in the list of processes, expect 1 result, got %d.", testLen)
	}

	// check to see if we can match our pid and our parent in the list
	the_check.Args = []string{"cmd", "-M", "-m"}
	pc.Init(the_check)

	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 2 != testLen {
		t.Errorf("Failed to include myself and my parent in the list of processes, expect 2 results, got %d.", testLen)
	}

	// check to see if we can match our processes command
	the_check.Args = []string{"cmd", "-m", "-M", "-p", "mo+"}
	pc.Init(the_check)

	testList = pc.excludeProcesses(matchlist)
	testLen = len(testList)
	if 1 != testLen {
		t.Errorf("Failed to exclude a process by regex, expect 1 result, got %d.", testLen)
	}

}

func getProc() process {
	return process{
		pid:         os.Getpid(),
		ppid:        os.Getppid(),
		command:     "moo --is --you",
		vsz:         33912,
		rss:         4164,
		pcpu:        10.0,
		threads:     65,
		cpuTime:     99999,
		processTime: 100,
		state:       "S",
		user:        "root",
	}
}

func getParentProc() process {
	return process{
		pid:         os.Getppid(),
		ppid:        1,
		command:     "yar --is --you",
		vsz:         33912,
		rss:         4164,
		pcpu:        0.0,
		threads:     65,
		cpuTime:     99999,
		processTime: 100,
		state:       "S",
		user:        "root",
	}
}
