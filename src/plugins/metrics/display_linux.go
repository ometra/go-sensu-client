package metrics

import (
	"fmt"
	"io/ioutil"
	"log"
	"plugins"
	"strings"
)

// PLATFORMS
//   Linux

func (display *DisplayStats) createPayload(r *plugins.Result) error {
	if !display.continue_gathering {
		return nil
	}
	content, err := ioutil.ReadFile("/sys/class/switch/hdmi/state")
	if nil != err {
		log.Printf("Failed to read HDMI State. %s", err)
		display.continue_gathering = false
		return nil
	}

	value := strings.Trim(string(content), "\n ")

	r.Add(fmt.Sprintf("display.hdmi %s", value))

	return nil
}
