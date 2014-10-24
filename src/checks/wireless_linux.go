package checks

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// PLATFORMS
//   Linux

func (ws *WirelessStats) createPayload(short_name string, timestamp uint) (string, error) {
	var payload string

	// is it better to use `arp -i wlan0` for this?
	//files, err := filepath.Glob("/proc/*/net/*/wlan*/all_sta_info")
	//if nil != err {
	//	return payload, err
	//}

	//for _, file := range files {
	//	content, err := ioutil.ReadFile(file)
	//	if nil != err {
	//		continue
	//	}

	//}

	return payload, nil
}
