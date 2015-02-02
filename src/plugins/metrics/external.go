package metrics

import (
	"log"
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
	log.Printf("About to start command (%s): %s", em.name, em.command)
	r.SetNoWrapOutput()
	cmd := exec.Command(em.command)

	out, errOut := cmd.CombinedOutput()

	if nil == errOut {
		r.Add(em.name + " " + string(out))
	}

	return errOut
}

func (em *ExternalMetric) GetStatus() string {
	return ""
}
