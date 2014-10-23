package checks

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

// PLATFORMS
//   Linux

func (cpu *CpuStats) setup() error {

	// get the number of CPUs from: /sys/devices/system/cpu/
	online, err := ioutil.ReadFile("/sys/devices/system/cpu/present")
	cpu.cpu_count = 1
	if nil != err {
		log.Printf("Unable to determine number of CPUs. Intialising only 1 CPU")
	} else {
		online_bits := strings.Split(string(online), "-")
		if len(online_bits) != 2 {
			log.Printf("/sys/devices/system/cpu/present CPU count file malformed. Initialising only 1 CPU")
		} else {
			cpu.cpu_count, err = strconv.Atoi(strings.Trim(online_bits[1], "\n"))
			if nil != err {
				log.Printf("Failed converting CPU count. Initialising on 1 CPU. %s", err)
				cpu.cpu_count = 1
			} else {
				// /sys/devices/system/cpu/present is 0 based
				cpu.cpu_count++
			}
		}
	}
	cpu.frequency = make(map[int]int, cpu.cpu_count)

	return nil
}

func (cpu *CpuStats) createPayload(short_name string, timestamp uint) (string, error) {
	var payload string
	// now inject our data
	for i := 0; i < cpu.cpu_count; i++ {
		cpu.frequency[i] = 0
		// attempt to load the file
		content, err := ioutil.ReadFile(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/cpuinfo_cur_freq", i))
		if nil == err {
			// we have content!
			cpu.frequency[i], err = strconv.Atoi(strings.Trim(string(content), "\n"))
			if nil != err {
				log.Printf("Failed to convert '%s' to an int", string(content))
			}
		} else {
			return payload, err
			log.Printf("Could not get CPU Freq for CPU %d: %s", i, err)
		}

		payload += fmt.Sprintf("%s.cpu.cpu%d.frequency %d %d\n", short_name, i, cpu.frequency[i], timestamp)
	}
	payload += fmt.Sprintf("%s.cpu.cpu_count %d %d\n", short_name, cpu.cpu_count, timestamp)

	// now time to get the stats
	file, err := ioutil.ReadFile("/proc/stat")
	if nil != err {
		return payload, err
	}

	cpu_metrics := []string{"user", "nice", "system", "idle", "iowait", "irq", "softirq", "steal", "guest"}
	//other_metrics := []string{"ctxt", "processes", "procs_running", "procs_blocked", "btime", "intr"}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if len(fields[0]) >= 3 && "cpu" == fields[0][0:3] {
			name := fields[0]
			if name == "cpu" {
				name = "total"
			}

			for i, field := range cpu_metrics {
				payload += fmt.Sprintf("%s.cpu.%s.%s %s %d\n", short_name, name, field, fields[i+1], timestamp)
			}
		}
		switch fields[0] {
		case "ctxt", "processes", "procs_running", "procs_blocked", "btime", "intr":
			payload += fmt.Sprintf("%s.cpu.%s %s %d\n", short_name, fields[0], fields[1], timestamp)
		}
	}

	return payload, nil
}
