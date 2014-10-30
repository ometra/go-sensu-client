package metrics

import (
	"fmt"
	"io/ioutil"
	"plugins"
	"strings"
)

func (u *UptimeStats) createPayload(r *plugins.Result) error {
	content, err := ioutil.ReadFile("/proc/uptime")
	if nil != err {
		return err
	}

	uptime_idle := strings.Split(strings.Trim(string(content), " \n"), " ")
	r.Add(fmt.Sprintf("uptime %s", uptime_idle[0]))
	r.Add(fmt.Sprintf("uptime_idle %s", uptime_idle[1]))

	return nil
}
