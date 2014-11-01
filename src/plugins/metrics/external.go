package metrics

import (
	"plugins"
)

type ExternalMetric struct {
	command string
	name    string
}

func (em *ExternalMetric) Init(config plugins.PluginConfig) (string, error) {
	// make sure that the command exists?
	em.name = config.Name
	return em.name, nil
}

func (em *ExternalMetric) Gather(r *plugins.Result) error {
	return nil
}

func (em *ExternalMetric) GetStatus() string {
	return ""
}
