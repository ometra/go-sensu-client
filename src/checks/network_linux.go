package checks

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (iface *NetworkInterfaceStats) createPayload(short_name string, timestamp uint) (string, error) {
	// grab all the stats in /sys/class/net/<interface>/statistics/*
	base_path := "/sys/class/net/"
	var output string
	interface_err := filepath.Walk(base_path, func(interface_path string, interface_info os.FileInfo, err error) error {
		if base_path == interface_path {
			return nil // no need to read the base path
		}
		// get a list of all the files in this path
		statistics_path := fmt.Sprintf("%s/statistics/", interface_path)
		file_err := filepath.Walk(statistics_path, func(file_path string, file_info os.FileInfo, err error) error {
			if statistics_path == file_path {
				return nil
			}
			value, err := ioutil.ReadFile(file_path)
			if nil != err {
				return err
			}

			output += fmt.Sprintf("%s.interface.%s.%s %s %d\n", short_name, interface_info.Name(), file_info.Name(), strings.Trim(string(value), " \n\t"), timestamp)
			return nil
		})

		return file_err
	})

	return output, interface_err
}
