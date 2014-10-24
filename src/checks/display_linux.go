package checks

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

// PLATFORMS
//   Linux

func (display *DisplayStats) createPayload(short_name string, timestamp uint) (string, error) {
	var payload string
	if !display.continue_gathering {
		return payload, nil
	}
	content, err := ioutil.ReadFile("/sys/class/switch/hdmi/state")
	if nil != err {
		log.Printf("Failed to read HDMI State. %s", err)
		display.continue_gathering = false
		return payload, nil
	}

	value := strings.Trim(string(content), "\n ")

	payload = fmt.Sprintf("%s.display.hdmi %s %d\n", short_name, value, timestamp)

	return payload, nil
}
