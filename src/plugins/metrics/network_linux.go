package metrics

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugins"
	"strings"
)

func (iface *NetworkInterfaceStats) createPayload(r *plugins.Result) error {
	// grab all the stats in /sys/class/net/<interface>/statistics/*
	base_path := "/sys/class/net/"

	interface_err := filepath.Walk(base_path, func(interface_path string, interface_info os.FileInfo, err error) error {
		if base_path == interface_path {
			return nil // no need to read the base path
		}
		// get a list of all the files in this path
		interface_name := interface_path[len(base_path):]
			if iface.do_filter {
				if v, ok := iface.interfaces[interface_name]; !ok || !v {
					return nil
				}
			}

		statistics_path := fmt.Sprintf("%s/statistics/", interface_path)
		file_err := filepath.Walk(statistics_path, func(file_path string, file_info os.FileInfo, err error) error {
			if statistics_path == file_path {
				return nil
			}
			value, err := ioutil.ReadFile(file_path)
			if nil != err {
				return err
			}

			r.Add(fmt.Sprintf("interface.%s.%s %s", interface_info.Name(), file_info.Name(), strings.Trim(string(value), " \n\t")))
			return nil
		})

		return file_err
	})

	return interface_err
}
