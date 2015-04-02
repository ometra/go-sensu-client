package sensu

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    "plugins"
    "plugins/checks"
    "plugins/metrics"
    "strconv"
    "strings"
    "time"
    "regexp"
    "github.com/bitly/go-simplejson"
)

type PluginProcessor struct {
    q                            MessageQueuer
    config                       *Config
    jobs                         map[string]plugins.SensuPluginInterface
    jobsConfig                   map[string]plugins.PluginConfig
    close                        chan bool
    publishResultsChan           chan bool
    saveResultsChan              chan bool
    results                      chan ResultInterface
    logger                       *log.Logger
    statsCollecting              bool // whether or not to set off more jobs
    stopCollectingOnNoConnection bool // whether or not to stop collecting stats when the connection to RabbitMQ drops
    statStore                    string
}

// used to create a new processor instance.
func NewPluginProcessor(w io.Writer, statStore string) *PluginProcessor {
    proc := new(PluginProcessor)
    proc.jobsConfig = make(map[string]plugins.PluginConfig)
    proc.results = make(chan ResultInterface, 600) // queue of 600 buffered results
    proc.publishResultsChan = make(chan bool)
    proc.saveResultsChan = make(chan bool)
    proc.logger = log.New(w, "Plugin: ", log.LstdFlags)
    proc.statStore = statStore

    return proc
}

// takes the config as a json blob and turns it into our running config for each check/metric
func newCheckConfig(json interface{}) plugins.PluginConfig {
    var conf plugins.PluginConfig

    converted := json.(map[string]interface{})

    if command, ok := converted["command"]; ok {
        conf.Command, _ = command.(string)
    }

    if args, ok := converted["args"]; ok {
        conf.Args, _ = args.([]string)
    } else {
        conf.Args = strings.Split(conf.Command, " ")
    }

    if handlers, ok := converted["handlers"]; ok {
        conf.Handlers, _ = handlers.([]string)
    }

    conf.Interval = 15 // default 15 second interval
    if interval, ok := converted["interval"]; ok {
        switch t := interval.(type) {
            default:
            i, err := strconv.ParseInt(fmt.Sprintf("%s", t), 10, 64)
            if nil == err {
                conf.Interval = time.Duration(i)
            }
            case int8, int16, int, int32, int64, uint8, uint16, uint, uint32, uint64:
            conf.Interval = time.Duration(interval.(int64))
        }
    }

    conf.Standalone = true
    if standalone, ok := converted["standalone"]; ok {
        conf.Standalone, _ = standalone.(bool)
    }

    if conf_type, ok := converted["type"]; ok {
        conf.Type, _ = conf_type.(string)
    }

    return conf
}

// does the funky command line variable replacing stuff
func commandReplace(command string, walkingConfigStart *simplejson.Json) string {
    commandRe := regexp.MustCompile(":::(.*?):::")

    command_bytes := commandRe.ReplaceAllFunc([]byte(command), func(match_bytes []byte) []byte {
        // make sure the :::'s are removed... because they are in match_bytes
        match_string := strings.Trim(string(match_bytes), ":")

        // so we should have a string like: value.value|default
        values := strings.Split(match_string, "|");
        var default_value, replace_var string
        replace_var = values[0]
        if len(values) > 1 {
            default_value = values[1]
        }

        // walk the value.value through the config tree
        steps := strings.Split(replace_var, ".")
        var found bool = true;
        var ok bool;
        walkingConfig := walkingConfigStart
        for _, step := range steps {
            if nil == walkingConfig {
                found = false
                break
            }
            walkingConfig, ok = walkingConfig.CheckGet(step)
            if !ok {
                found = false;
                break
            }
        }

        if found {
            bytes, _ := walkingConfig.Bytes();
            return bytes;
        } else {
            return []byte(default_value);
        }
    })

    return string(command_bytes);
}

// helper function to add a check to the queue of checks
func (p *PluginProcessor) AddJob(job plugins.SensuPluginInterface, checkConfig plugins.PluginConfig) {
    name, err := job.Init(checkConfig)
    if nil != err {
        p.logger.Printf("Failed to initialise check: (%s) %s\n", name, err)
        return
    }

    checkConfig.Command = commandReplace(checkConfig.Command, p.config.Data().Get("client"))

    p.logger.Printf("Scheduling job: %s (%s) every %d seconds", name, checkConfig.Command, checkConfig.Interval)

    p.jobs[name] = job
    p.jobsConfig[name] = checkConfig
}

// called to set things up
func (p *PluginProcessor) Init(q MessageQueuer, config *Config) error {
    if err := q.ExchangeDeclare(
    RESULTS_QUEUE,
    "direct",
    ); err != nil {
        return fmt.Errorf("Exchange Declare: %s", err)
    }

    p.q = q
    p.config = config
    p.close = make(chan bool, len(p.jobs)+1)
    var check plugins.SensuPluginInterface

    // load the checks we want to do
    checks_config := p.config.Data().Get("checks").MustMap()
    p.jobs = make(map[string]plugins.SensuPluginInterface)

    for check_type, checkConfigInterface := range checks_config {
        checkConfig, ok := checkConfigInterface.(map[string]interface{})
        if !ok {
            p.logger.Printf("Failed to parse config: %", check_type)
            continue
        }

        config := newCheckConfig(checkConfig)
        check = getCheckHandler(check_type, config.Type)

        config.Name = check_type

        p.AddJob(check, config)
    }

    return nil
}

// gets the Gather of checks/metrics going
func (p *PluginProcessor) Start() {
    go p.publishResults()
    if p.statsCollecting {
        // since Start() gets called when we have a good Rabbit connection - we can stop storing our results in a file
        p.saveResultsChan <- false
        return
    }

    // we are collecting results now - used so that we do not fire up a second copy of the stats gathering
    p.statsCollecting = true

    clientConfig := p.config.Client

    // start our result publisher thread
    for job_name, job := range p.jobs {

        // this is the main stats gathering function
        go func(theJobName string, theJob plugins.SensuPluginInterface) {
            config := p.jobsConfig[theJobName]

            reset := make(chan bool)

            timer := time.AfterFunc(0, func() {
                p.logger.Printf("Gathering: %s", theJobName)
                result := NewResult(clientConfig, theJobName)
                result.SetCommand(config.Command)

                plugin_result := new(plugins.Result)

                err := theJob.Gather(plugin_result)
                result.SetWrapOutput(!plugin_result.IsNoWrapOutput()) // for external checks
                result.SetOutput(plugin_result.Output())
                result.SetCheckStatus(theJob.GetStatus())

                if nil != err {
                    // returned an error - we should stop this job from running
                    p.logger.Printf("Failed to gather stat: %s. %v", theJobName, err)
                    reset <- false
                    return
                }

                // add it to the processing queue
                p.results <- result

                reset <- true
            })

            defer timer.Stop()
            for {
                select {
                case cont := <-reset:
                    if cont {
                        timer.Reset(config.Interval * time.Second)
                    } else {
                        timer.Stop()
                    }
                case <-p.close: // shutting down stats gather message
                    return
                }
            }
        }(job_name, job)
    }
}

// Puts a halt to all of our checks/metrics gathering
func (p *PluginProcessor) Stop(force bool) {

    // we *could* stop the automated stat gathering here by sending close messages
    // but we have found that gathering stats while the rabbitmq connection is broken
    // to be rather handy
    if p.stopCollectingOnNoConnection || force {
        p.logger.Printf("STOP: Closing %d Plugins: ", len(p.jobs))
        p.statsCollecting = false
        for name, _ := range p.jobs {
            p.logger.Print("STOP: Closing Plugin: ", name)
            p.close <- true
        }
        p.publishResultsChan <- false
    } else {
        // tell our result publishing to stop.
        p.publishResultsChan <- true
    }
}

func (p *PluginProcessor) loadResults() {
    // get the results from file
    f, err := os.OpenFile(p.statStore, os.O_RDWR|os.O_EXCL, 0600)
    if err != nil {
        return
    }

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        sr := new(SavedResult)
        sr.SetResult(scanner.Text())
        p.results <- sr
    }

    // once we have the contents of the file sent back to the result queue, truncate the file!
    f.Truncate(0)
    f.Close()
}

// instead of writing the stats to RabbitMQ (i.e. rabbit connection has gone away)
// we write them to a file instead, so that we may send them on once the connection
// to rabbit has been reestablished
func (p *PluginProcessor) saveResults() {
    // does the stat store file exist?
    p.logger.Printf("START: Disk store (%s) for results...", p.statStore)
    var f *os.File
    for {
        select {
        case result := <-p.results:
            if result.HasOutput() {
                finfo, err := os.Stat(p.statStore) // is there a file here?
                if err != nil {
                    // no file? create one!
                    f, err = os.OpenFile(p.statStore, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
                } else {
                    // oh, yes, let's open it (if we do not have too much in there already)
                    if finfo.Size() > 104857600 {
                        err = fmt.Errorf("The stat file is too large, discarding stat")
                    } else {
                        f, err = os.OpenFile(p.statStore, os.O_APPEND|os.O_WRONLY|os.O_EXCL, 0600)
                    }
                }

                if err != nil {
                    p.logger.Println("Cannot write to stat store,", err)

                } else {
                    f.Write(result.toJson())
                    f.WriteString("\n")
                    f.Close()
                }
            }
        case <-p.saveResultsChan:
            p.logger.Println("STOP: Result saving to file...")
            go p.publishResults()
            return
        }
    }
}

// our result publishing. will publish results until we call PluginProcessor.Stop()
func (p *PluginProcessor) publishResults() {
    go p.loadResults()
    p.logger.Println("START: Result publishing to RabbitMQ...")
    for {
        //p.logger.Printf("Result Queue State: %d/%d\n", len(p.results), cap(p.results))
        select {
        case result := <-p.results:
            if result.HasOutput() {
                if err := p.q.Publish(RESULTS_QUEUE, "", result.GetPayload()); err != nil {
                    p.logger.Printf("Error Publishing Stats: %v.", err)
                    p.results <- result // requeue the failed result
                }
            }
        case cont := <-p.publishResultsChan:
            p.logger.Print("STOP: Shutting down result publishing to RabbitMQ")
            if cont {
                go p.saveResults()
            }
            return
        }
    }
}

// determines if we can use one of our internet plugins to handle the check.
// if not, it will use an external check
func getCheckHandler(check_type, config_type string) plugins.SensuPluginInterface {
    var check plugins.SensuPluginInterface

    check = plugins.GetPlugin(check_type)
    if check == nil {
        if "metric" == config_type {
            // we have a metric!
            check = new(metrics.ExternalMetric)
        } else {
            // we have a check!
            check = new(checks.ExternalCheck)
        }
    }

    return check
}
