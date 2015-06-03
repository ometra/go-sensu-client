package metrics

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"plugins"
	"regexp"
	"strings"
	"time"
)

// TCP Network stats
//
// gobs of this code lifted from: https://github.com/grahamking/latency/
//
// DESCRIPTION
//  This plugin attempts to determine network latency.
// interface is up and we cannot ping.
//
// OUTPUT
//   Graphite plain-text format (name value timestamp\n)
//
// PLATFORMS
//   Linux

const TCP_STATS_NAME = "tcp_metrics"

type TcpStats struct {
	flags            *flag.FlagSet
	networkInterface string
	listenInterface  string
	remoteAddress    string
	localAddress     string
	networkPort      int
	timeout          float64
	workingTimeout   time.Duration
	retryCount       int
	hostNiceName     string
}

type receiveErrorType struct {
	err string
}

func (re receiveErrorType) Error() string {
	return re.err
}
func init() {
	plugins.Register("tcp_metrics", new(TcpStats))
}

func (tcp *TcpStats) Init(config plugins.PluginConfig) (string, error) {
	tcp.flags = flag.NewFlagSet("tcp-metrics", flag.ContinueOnError)

	tcp.flags.StringVar(&tcp.networkInterface, "test-interface", "", "The Network to test before pinging, defaults to the listen interface")
	tcp.flags.StringVar(&tcp.listenInterface, "i", "", "The network interface to listen on")
	tcp.flags.StringVar(&tcp.remoteAddress, "host", "", "The Network Address to ping")
	tcp.flags.IntVar(&tcp.networkPort, "port", 22, "The Port to SYN (Ping)")
	tcp.flags.Float64Var(&tcp.timeout, "timeout", 10, "Number of seconds to wait for a response")
	tcp.flags.IntVar(&tcp.retryCount, "retry-count", 3, "The number of times to retry before failing")

	var err error
	if len(config.Args) > 1 {
		err = tcp.flags.Parse(config.Args[1:])
		if nil != err {
			return TCP_STATS_NAME, err
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

	tcp.workingTimeout, err = time.ParseDuration(fmt.Sprintf("%0.0fms", tcp.timeout*1000))
	if err != nil {
		log.Println(err)
	}

	r := regexp.MustCompile("[^0-9a-zA-Z]")
	tcp.hostNiceName = r.ReplaceAllString(tcp.remoteAddress, "_")

	if "" != os.Getenv("DEBUG") {
		log.Println("Remote Host:     ", tcp.remoteAddress)
		log.Println("Listen Interface:", tcp.listenInterface)
		log.Println("Test interface:  ", tcp.networkInterface)
		log.Println("Port:            ", tcp.networkPort)
		log.Println("Retry count:     ", tcp.retryCount)
		log.Printf("Ping Timeout:     %s", tcp.workingTimeout.String())
	}

	return TCP_STATS_NAME, err
}

func (tcp *TcpStats) Gather(r *plugins.Result) error {
	// measure TCP/IP response

	stat, err := os.Stat("/sys/class/net/" + tcp.networkInterface)
	if nil != err {
		return fmt.Errorf("Interface %s does not exist.", tcp.networkInterface)
	}

	if !stat.IsDir() {
		return fmt.Errorf("Interface %s does not exist.", tcp.networkInterface)
	}

	// is the network interface up?
	state, err := ioutil.ReadFile("/sys/class/net/" + tcp.networkInterface + "/operstate")
	if nil != err {
		return fmt.Errorf("Unable to determine if interface is up.")
	}
	// cannot ping when the network is down
	if "up" != string(state[0:2]) {
		return fmt.Errorf("Network Interface %s is down", tcp.networkInterface)
	}

	iface, err := interfaceAddress(tcp.listenInterface)
	if err != nil {
		log.Print(err)
		// we do not return the error, because that will cause the check to be stopped.
		// we return nil and no stats instead while we wait for the interface to get an
		// ip address again. (e.g. happens when network manager disables interface)
		return nil
	} else {
		tcp.localAddress = strings.Split(iface.String(), "/")[0]
	}

	// does the remoteAddress look like an IP address?
	remoteIp, err := getRemoteAddress(tcp.remoteAddress)
	if err != nil {
		return err
	}

	var counter int
	var totalLatency time.Duration
	if "" != tcp.localAddress {
	TryLoop:
		for counter < tcp.retryCount {
			counter++
			latency, errPing := tcp.ping(tcp.localAddress, remoteIp, uint16(tcp.networkPort))
			if errPing == nil {
				totalLatency += latency
				r.Add(fmt.Sprintf("tcp.latency.%s.ms %0.2f", tcp.hostNiceName, float32(totalLatency)/float32(time.Millisecond)))
				r.Add(fmt.Sprintf("tcp.try-count.%s %d", tcp.hostNiceName, counter))
				break
			}
			switch errPing.(type) {
			case receiveErrorType:
				//log.Println(errPing)
				break TryLoop
			case error:
				totalLatency += tcp.workingTimeout
				log.Printf("Failed TCP Ping check %d...", counter)
			}
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

func (tcp *TcpStats) ping(localAddr, remoteAddr string, port uint16) (time.Duration, error) {
	receiveDuration := make(chan time.Duration)
	receiveError := make(chan error)
	timeoutChannel := make(chan bool)

	// limit ourselves to 10 seconds
	time.AfterFunc(tcp.workingTimeout, func() { timeoutChannel <- true })

	go func() {
		t, err := latency(localAddr, remoteAddr, port)
		if err != nil {
			receiveError <- err
		} else {
			receiveDuration <- t
		}
	}()

	select {
	case d := <-receiveDuration:
		return d, nil
	case e := <-receiveError:
		var re receiveErrorType
		re.err = e.Error()
		return 0, re
	case <-timeoutChannel:
		return time.Duration(0), fmt.Errorf("Failed to TCP ping remote host")
	}
}
