package checks

import (
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
	em.command = config.Command
	return ec.name, nil
}

func (ec *ExternalCheck) Gather(r *plugins.Result) error {
	cmd := exec.Command(em.command)
	err := cmd.Run()
	ec.checkStatus = plugins.UNKNOWN

	out, errOut := cmd.CombinedOutput()
	r.SetOutput(string(out))
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
