package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sensu"
	"strings"
	"syscall"
)

var (
	configFile, configDir string
	statStoreFile         string
	logOutput             io.Writer = os.Stdout
	overrideHostName      string
	overrideAddress       string
	quiet                 bool
)

type QuietWriter struct{}

func (q QuietWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func init() {
	flag.StringVar(&configFile, "config-file", "config.json", "Sensu JSON config file")
	flag.StringVar(&configDir, "config-dir", "conf.d", "directory or comma-delimited directory list for Sensu JSON config files")
	flag.StringVar(&statStoreFile, "stat-store", "", "The file to store results when we cannot get a RabbitMQ connection. defaults to the same location as config-file/config-dir")
	flag.StringVar(&overrideHostName, "hostname", "", "A host name to use instead of the one found in the config")
	flag.StringVar(&overrideAddress, "address", "", "An Address to override the one found in the config file")
	flag.BoolVar(&quiet, "quiet", false, "When true makes all logger output go to dev null")
	flag.Parse()
}

func runner(stop chan bool) {
	configDirs := strings.Split(configDir, ",")
	settings, err := sensu.LoadConfigs(configFile, configDirs)

	if err != nil {
		log.Printf("Unable to load settings: %s", err)
		flag.Usage()
		os.Exit(1)
	}

	if "" != overrideHostName {
		settings.Client.Name = overrideHostName
		settings.Data().Get("client").Set("name", overrideHostName)
	}

	if "" != overrideAddress {
		settings.Client.Address = overrideAddress
		settings.Data().Get("client").Set("address", overrideAddress)
	}

	if "" == statStoreFile {
		// put it in the same place as the config
		var path string
		if "" != configFile {
			path, _ = filepath.Abs(filepath.Dir(configFile))
		}

		statStoreFile = path + "/stat.store"
	}

	processes := []sensu.Processor{
		sensu.NewKeepalive(logOutput),
		sensu.NewSubscriber(logOutput),
		sensu.NewPluginProcessor(logOutput, statStoreFile),
	}
	c := sensu.NewClient(settings, processes)

	c.Start(stop)
	stop <- true
}

func main() {
	if quiet {
		logOutput = QuietWriter{}
		log.SetOutput(QuietWriter{})
	}

	osSignalChan := make(chan os.Signal, 3)
	signal.Notify(osSignalChan, os.Interrupt, os.Kill, syscall.SIGUSR1)

	stop := make(chan bool)
	run := make(chan bool, 1)
	run <- true

	interrupt_count := 0

	for {

		select {
		case <-run:
			go runner(stop)

		case sig := <-osSignalChan:
			log.Println("Got a signal! ", sig)
			switch sig {
			case os.Interrupt, os.Kill:
				interrupt_count++
				if interrupt_count > 1 {
					log.Println("That's the second interrupt - forcing exit!")
					os.Exit(0)
				}
				log.Println("Closing all the things!")
				stop <- true
				log.Println("... waiting")
				<-stop
				return
			default:
				log.Println("Reloading our Config")
				// reload our config!
				stop <- true
				<-stop
				run <- true
			}
		}
	}
}
