package checks

import (
	"plugins"
)

type ExternalCheck struct {
	command string
	name    string
}

func (ec *ExternalCheck) Init(config plugins.PluginConfig) (string, error) {
	// make sure that the command exists?
	ec.name = config.Name
	return ec.name, nil
}

func (ec *ExternalCheck) Gather(r *plugins.Result) error {
	return nil
}

func (ec *ExternalCheck) GetStatus() string {
	return ec.name + " " + ec.checkStatus.ToString()
}
