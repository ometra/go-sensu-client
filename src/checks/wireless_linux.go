package checks

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PLATFORMS
//   Linux

func (ws *WirelessStats) setup() error {

	// this grabs the clients connected to the AP for realtek devices

	matches, err := filepath.Glob("/proc/net/rtl*/wlan*/all_sta_info")
	if nil != err {
		return err
	}

	if nil == matches {
		return fmt.Errorf("Cannot get connected wireless devices")
	}

	ws.files = matches

	// now get the excludes
	excludes, err := filepath.Glob("/sys/class/net/*/address")
	if nil != err {
		return err
	}

	if nil == excludes {
		return fmt.Errorf("This device has no network interfaces")
	}

	ws.exclude = make([]string, len(excludes)+1)

	for i, file := range excludes {
		content, err := ioutil.ReadFile(file)
		if nil != err {
			continue
		}
		ws.exclude[i] = strings.Trim(string(content), " \n")
	}
	ws.exclude[len(excludes)] = "ff:ff:ff:ff:ff:ff"

	return nil
}

func (ws *WirelessStats) createPayload(short_name string, timestamp uint) (string, error) {
	var payload string
	var counter int
	var skip bool

	for _, file := range ws.files {
		file, err := os.Open(file)
		if nil != err {
			log.Printf("Unable to open file (%s): %s", file, err)
			continue
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			skip = false

			if !strings.Contains(line, "sta's macaddr") {
				continue
			}

			for _, exclude := range ws.exclude {
				if strings.Contains(line, exclude) {
					skip = true
				}
			}
			if skip {
				continue
			}
			counter++
		}
	}

	payload = fmt.Sprintf("%s.access_point.connected_clients %d %d\n", short_name, counter, timestamp)

	return payload, nil
}
