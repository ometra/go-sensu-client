package checks

import (
	"io/ioutil"
	"strings"
	"fmt"
)

// PLATFORMS
//   Linux

func (load *LoadStats) createLoadAveragePayload(short_name string, timestamp uint) (string, error) {
	var payload string
	content, err := ioutil.ReadFile("/proc/loadavg")
	if nil != err {
		return payload, err
	}

	bits := strings.Split(string(content), " ")

	payload = fmt.Sprintf("%s.load_avg.one %s %d\n", short_name, bits[0], timestamp)
	payload += fmt.Sprintf("%s.load_avg.five %s %d\n", short_name, bits[1], timestamp)
	payload += fmt.Sprintf("%s.load_avg.fifteen %s %d\n", short_name, bits[2], timestamp)

	return payload, nil
}
