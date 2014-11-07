package checks

import (
	"fmt"
	"os/exec"
	"plugins"
)

type ExternalCheck struct {
	command     string
	name        string
	checkStatus plugins.Status
}

func (ec *ExternalCheck) Init(config plugins.PluginConfig) (string, error) {
	// make sure that the command exists?
	ec.name = config.Name
	ec.command = config.Command
	return ec.name, nil
}

func (ec *ExternalCheck) Gather(r *plugins.Result) error {
	fmt.Printf("About to start command\n")
	cmd := exec.Command(ec.command)
	err := cmd.Run()
	ec.checkStatus = plugins.UNKNOWN

	out, errOut := cmd.CombinedOutput()
	fmt.Println("Output BELOW")
	fmt.Println(out)
	r.Add(string(out))
	if nil == errOut {
		ec.checkStatus = plugins.OK
	} else {
		ec.checkStatus = plugins.WARNING
	}

	return err
}

func (ec *ExternalCheck) GetStatus() string {
	return ec.name + " " + ec.checkStatus.ToString()
}
