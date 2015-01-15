package main

import (
	"flag"
	"io"
	"log"
	"os"
	"sensu"
	"strings"
)

var (
	configFile, configDir string
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
	flag.StringVar(&overrideHostName, "hostname", "", "A host name to use instead of the one found in the config")
	flag.StringVar(&overrideAddress, "address", "", "An Address to override the one found in the config file")
	flag.BoolVar(&quiet, "quiet", false, "This makes all logger output go to dev null")
	flag.Parse()
}

func main() {
	if quiet {
		logOutput = QuietWriter{}
		log.SetOutput(QuietWriter{})
	}

	configDirs := strings.Split(configDir, ",")
	settings, err := sensu.LoadConfigs(configFile, configDirs)
	if "" != overrideHostName {
		settings.Client.Name = overrideHostName
	}

	if "" != overrideAddress {
		settings.Client.Address = overrideAddress
	}

	if err != nil {
		log.Printf("Unable to load settings: %s", err)
		flag.Usage()
		os.Exit(1)
	}

	processes := []sensu.Processor{
		sensu.NewKeepalive(logOutput),
		sensu.NewSubscriber(logOutput),
		sensu.NewPluginProcessor(logOutput),
	}
	c := sensu.NewClient(settings, processes)

	c.Start()
}
