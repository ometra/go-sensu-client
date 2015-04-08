package metrics

import (
	"flag"
	"log"
	"plugins"
	"strings"
)

// network interfaces stats for Linux based machines
//
// DESCRIPTION
//  This plugin gets the interface metrics
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

type NetworkInterfaceStats struct {
	flags          *flag.FlagSet
	req_interfaces string
	interfaces     map[string]bool
	do_filter      bool
}

func init() {
	plugins.Register("interface_metrics", new(NetworkInterfaceStats))
}

func (iface *NetworkInterfaceStats) Init(config plugins.PluginConfig) (string, error) {
	iface.flags = flag.NewFlagSet("tcp-metrics", flag.ContinueOnError)

	iface.flags.StringVar(&iface.req_interfaces, "i", "", "The list of interfaces to include, all others excluded")

	var err error
	if len(config.Args) > 1 {
		err = iface.flags.Parse(config.Args[1:])
		if nil != err {
			log.Printf("Failed to parse process check command line: %s", err)
		}
	}

	iface.interfaces = make(map[string]bool)
	if "" != iface.req_interfaces {
		iface.do_filter = true
		for _, key := range strings.Split(iface.req_interfaces, ",") {
			iface.interfaces[strings.Trim(key, " ")] = true
		}
	}

	return "interface_metrics", nil
}

func (iface *NetworkInterfaceStats) Gather(r *plugins.Result) error {
	return iface.createPayload(r)
}

func (iface *NetworkInterfaceStats) GetStatus() string {
	return ""
}
