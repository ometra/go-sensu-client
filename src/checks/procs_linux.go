package checks

import (
	"fmt"
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

	if pc.metric > 0 {
		count := len(process_list)
		payload += fmt.Sprintf("; %d == %d", pc.metric, count)
	} else {
		count := len(process_list)
	}

	return payload, nil
}

func (pc *ProcessCheck) gatherProcesses() []process {

}

func (pc *ProcessCheck) excludeProcesses(list []process) []process {
}
