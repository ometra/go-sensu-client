package metrics

import (
	"fmt"
	"io/ioutil"
	"log"
	"plugins"
	"regexp"
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

	cpu.gather_frequency_stats = true

	return nil
}

func (cpu *CpuStats) getCpuValue(file string) uint64 {
	file_content, err := ioutil.ReadFile(file)
	content := strings.Trim(string(file_content), "\n ")

	var value uint64 = 0
	if nil == err {
		// we have content!
		value, err = strconv.ParseUint(content, 10, 64)
		if nil != err {
			cpu.failed_freq_gather_count++
			log.Printf("Failed to convert '%s' to an int", content)
		}
	} else {
		cpu.failed_freq_gather_count++
		log.Printf("Unable to read file: %s. %s", file, err)
	}
	return value
}

func (cpu *CpuStats) createPayload(r *plugins.Result) error {
	var speed uint64

	if cpu.gather_frequency_stats {
		// grab our frequency stats
		for i := 0; i < cpu.cpu_count; i++ {
			speed = cpu.getCpuValue(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/cpuinfo_cur_freq", i))
			r.Add(fmt.Sprintf("cpu.cpu%d.frequency.current %d", i, speed))

			speed = cpu.getCpuValue(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/cpuinfo_max_freq", i))
			r.Add(fmt.Sprintf("cpu.cpu%d.frequency.max %d", i, speed))

			speed = cpu.getCpuValue(fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/cpuinfo_min_freq", i))
			r.Add(fmt.Sprintf("cpu.cpu%d.frequency.min %d", i, speed))
		}

		if cpu.failed_freq_gather_count >= cpu.cpu_count {
			cpu.gather_frequency_stats = false
			log.Printf("Failed gathering CPU Frequency Stats. Disabling future freq gathering.")
		}
	}
	r.Add(fmt.Sprintf("cpu.cpu_count %d", cpu.cpu_count))

	// now time to get the CPU stats
	file, err := ioutil.ReadFile("/proc/stat")
	if nil != err {
		return err
	}

	cpu_metrics := []string{"user", "nice", "system", "idle", "iowait", "irq", "softirq", "steal", "guest"}

	regex, err := regexp.Compile("\\s+")
	if nil != err {
		log.Printf("Failed to compile regex! %s", err)
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		fields := regex.Split(line, 12)
		if len(fields[0]) >= 3 && "cpu" == fields[0][0:3] {
			name := fields[0]
			if name == "cpu" {
				name = "total"
			}

			for i, field := range cpu_metrics {
				r.Add(fmt.Sprintf("cpu.%s.%s %s", name, field, fields[i+1]))
			}
		}
		switch fields[0] {
		case "ctxt", "processes", "procs_running", "procs_blocked", "btime", "intr":
			r.Add(fmt.Sprintf("cpu.%s %s", fields[0], fields[1]))
		}
	}

	return nil
}
