package metrics

import (
	"fmt"
	"io/ioutil"
	"plugins"
	"regexp"
	"strconv"
	"strings"
)

var memoryKeys = map[string]string{
	"MemTotal":  "total",
	"MemFree":   "free",
	"Buffers":   "buffers",
	"Cached":    "cached",
	"SwapTotal": "swapTotal",
	"SwapFree":  "swapFree",
	"Dirty":     "dirty",
}

func (mem *MemoryStats) createPayload(r *plugins.Result) error {
	file, err := ioutil.ReadFile("/proc/meminfo")
	memoryValues := make(map[string]int64)

	if nil != err {
		return err
	}
	re, err := regexp.Compile("[\\s\\:]+")
	if nil != err {
		return err
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		parts := re.Split(line, 3)
		if label, ok := memoryKeys[parts[0]]; ok {
			memoryValues[label], err = strconv.ParseInt(parts[1], 10, 64)
			if nil != err {
				memoryValues[label] = int64(0)
			}
		}
	}

	// some additional values
	memoryValues["swapUsed"] = memoryValues["swapTotal"] - memoryValues["swapFree"]
	memoryValues["used"] = memoryValues["total"] - memoryValues["free"]
	memoryValues["usedWOBuffersCaches"] = memoryValues["used"] - (memoryValues["buffers"] + memoryValues["cached"])
	memoryValues["freeWOBuffersCaches"] = memoryValues["free"] + (memoryValues["buffers"] + memoryValues["cached"])
	if memoryValues["swapTotal"] > 0 {
		memoryValues["swapUsedPercentage"] = 100 * memoryValues["swapUsed"] / memoryValues["swapTotal"]
	}

	for label, value := range memoryValues {
		// memory is reported in KB, we need Bytes - bitshift 10 is the same as *1024
		r.Add(fmt.Sprintf("memory.%s %d", label, value<<10)) // value<<10 == value*1024
	}

	return nil
}
