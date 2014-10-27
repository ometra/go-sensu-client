package checks

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func (u *UptimeStats) createPayload(short_name string, timestamp uint) (string, error) {
	var payload string
	content, err := ioutil.ReadFile("/proc/uptime")
	if nil != err {
		return payload, err
	}

	uptime_idle := strings.Split(strings.Trim(string(content), " \n"), " ")
	payload = fmt.Sprintf("%s.uptime %s %d\n", short_name, uptime_idle[0], timestamp)
	payload += fmt.Sprintf("%s.uptime_idle %s %d\n", short_name, uptime_idle[1], timestamp)

	return payload, nil
}
