package metrics

import (
	"fmt"
	"io/ioutil"
	"plugins"
	"strings"
)

// PLATFORMS
//   Linux

func (load *LoadStats) createPayload(r *plugins.Result) error {
	content, err := ioutil.ReadFile("/proc/loadavg")
	if nil != err {
		return err
	}

	bits := strings.Split(string(content), " ")

	r.Add(fmt.Sprintf("load_avg.one %s", bits[0]))
	r.Add(fmt.Sprintf("load_avg.five %s", bits[1]))
	r.Add(fmt.Sprintf("load_avg.fifteen %s", bits[2]))

	return nil
}
