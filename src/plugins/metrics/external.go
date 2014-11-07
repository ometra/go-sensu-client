package metrics

import (
	"fmt"
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
	fmt.Printf("About to start command\n")
	cmd := exec.Command(em.command)
	err := cmd.Run()

	out, errOut := cmd.CombinedOutput()
	fmt.Println("Output BELOW")
	fmt.Println(out)

	if nil == errOut {
		r.Add(em.name + " " + string(out))
	}

	return err
}

func (em *ExternalMetric) GetStatus() string {
	return ""
}
