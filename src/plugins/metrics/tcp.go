package metrics

import (
	"plugins"
	"flag"
	"log"
	"net"
	"time"
	"strings"
	"io/ioutil"
	"fmt"
	"os"
	"strconv"
	"regexp"
)

// TCP Network stats
//
// gobs of this code lifted from: https://github.com/grahamking/latency/
//
// DESCRIPTION
//  This plugin attempts to determine network latency. It can optionally run a command when the specified
// interface is up and we cannot ping
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

const TCP_STATS_NAME = "tcp_metrics"

type TcpStats struct {
	flags            *flag.FlagSet
	networkInterface      string
	listenInterface       string
	remoteAddress         string
	localAddress          string
	networkPort           int
	timeout               int
	workingTimeout        time.Duration
	reboot                bool
	rebootStatFile        string
	hostNiceName          string
}

func (tcp *TcpStats) Init(config plugins.PluginConfig) (string, error) {
	tcp.flags = flag.NewFlagSet("tcp-metrics", flag.ContinueOnError)

	tcp.flags.StringVar(&tcp.networkInterface, "test-interface", "", "The Network to test before pinging, defaults to the listen interface")
	tcp.flags.StringVar(&tcp.listenInterface, "i", "", "The network interface to listen on")
	tcp.flags.StringVar(&tcp.remoteAddress, "host", "", "The Network Address to ping")
	tcp.flags.IntVar(&tcp.networkPort, "port", 22, "The Port to SYN (Ping)")
	tcp.flags.IntVar(&tcp.timeout, "timeout", 10, "Number of seconds to wait for a response")
	tcp.flags.BoolVar(&tcp.reboot, "reboot", false, "If the network is up and ping does not work - reboot")
	tcp.flags.StringVar(&tcp.rebootStatFile, "reboot-stat-file", "", "If specified this file is written to before the reboot action and a system.reboot.tcp counter sent after reboot")

	var err error
	if len(config.Args) > 1 {
		err = tcp.flags.Parse(config.Args[1:])
		if nil != err {
			log.Printf("Failed to parse process check command line: %s", err)
		}
	}

	if "" == tcp.listenInterface {
		return TCP_STATS_NAME, fmt.Errorf("You need to specify an Interface! e.g.: -i eth0")
	}

	if "" == tcp.networkInterface {
		tcp.networkInterface = tcp.listenInterface
	}

	if "" == tcp.remoteAddress {
		return TCP_STATS_NAME, fmt.Errorf("You need to specify a host to ping! e.g.: -host 10.0.0.1")
	}

	tcp.workingTimeout = time.Duration(tcp.timeout)*time.Second

	r := regexp.MustCompile("[^0-9a-zA-Z]")
	tcp.hostNiceName = r.ReplaceAllString(tcp.remoteAddress, "_")

	return TCP_STATS_NAME, err
}

func (tcp *TcpStats) Gather(r *plugins.Result) error {
	// measure TCP/IP response
	var rebootCount uint

	// is the network interface up?
	state, err := ioutil.ReadFile("/sys/class/net/" + tcp.networkInterface + "/operstate")
	if nil != err {
		return err
	}
	// cannot ping when the network is down
	if "up" != string(state[0:2]) {
		return nil
	}

	iface, err := interfaceAddress(tcp.listenInterface)
	if err != nil {
		log.Print(err)
		return nil
	}
	tcp.localAddress = strings.Split(iface.String(), "/")[0]

	addrs, err := net.LookupHost(tcp.remoteAddress)
	if err != nil {
		log.Printf("Error resolving %s. %s\n", tcp.remoteAddress, err)
	}

	if "" != tcp.rebootStatFile {
		rebootCount = tcp.getRebootCount()
		r.Add(fmt.Sprintf("tcp.reboot-count %d", rebootCount))
		tcp.setRebootCount(uint(0))
	}

	latency, errPing := tcp.ping(tcp.localAddress, addrs[0], uint16(tcp.networkPort))
	if errPing == nil {
		r.Add(fmt.Sprintf("tcp.latency.%s.ms %0.2f", tcp.hostNiceName, float32(latency)/float32(time.Millisecond)))
	} else {

		if tcp.reboot {
			// if we have a file to write to, write a reboot counter
			if "" != tcp.rebootStatFile {
				rebootCount++
				tcp.setRebootCount(rebootCount)
				r.Add(fmt.Sprintf("tcp.reboot-count %d", rebootCount))
			}
			log.Println("TCP Check Failed - Rebooting the system in 2 seconds")
			// this gives the system time to flush

			time.AfterFunc(2*time.Second, func() {
					tcp.performReboot()
				})
		}
	}

	return nil
}

func (tcp *TcpStats) GetStatus() string {
	return ""
}

func (tcp *TcpStats) ShowUsage() {
	tcp.flags.PrintDefaults()
}

func (tcp *TcpStats) getRebootCount() uint {
	var count uint
	content, err := ioutil.ReadFile(tcp.rebootStatFile)
	if err != nil {
		return 0
	}
	// we have an existing count!
	c, err := strconv.ParseUint(string(content), 10, 32)
	if err != nil {
		c = 0
	}
	count = uint(c)

	return count
}

func (tcp *TcpStats) setRebootCount(count uint) {
	value := []byte(fmt.Sprintf("%d", count))
	err := ioutil.WriteFile(tcp.rebootStatFile, value, os.FileMode(0644))
	if err != nil {
		log.Println(err)
	}
}

func (tcp *TcpStats) ping(localAddr, remoteAddr string, port uint16) (time.Duration, error) {
	receiveDuration := make(chan time.Duration)
	timeoutChannel := make(chan bool)

	// limit ourselves to 10 seconds
	time.AfterFunc(tcp.workingTimeout, func() { timeoutChannel <- true })

	go func() {
		receiveDuration <- latency(localAddr, remoteAddr, port)
	}()

	select {
	case d := <-receiveDuration:
		return d, nil
	case <-timeoutChannel:
		return time.Duration(0), fmt.Errorf("Failed to TCP ping remote host")
	}
}
