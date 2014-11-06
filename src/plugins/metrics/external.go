package metrics

import (
	"os/exec"
	"plugins"
)

type ExternalMetric struct {
	command string
	name    string
}

func (em *ExternalMetric) Init(config plugins.PluginConfig) (string, error) {
	// make sure that the command exists?
	em.name = config.Name
	em.command = config.Command
	return em.name, nil
}

func (em *ExternalMetric) Gather(r *plugins.Result) error {
	cmd := exec.Command(em.command)
	err := cmd.Run()

	out, errOut := cmd.CombinedOutput()
	if nil == errOut {
		r.SetOutput(string(bytes))
	}

	return err
}

func (em *ExternalMetric) GetStatus() string {
	return ""
}
