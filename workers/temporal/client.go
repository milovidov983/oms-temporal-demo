package temporal

import (
	"log"

	"go.temporal.io/sdk/client"
)

type TemporalClientConfig struct {
	HostPort  string
	Namespace string
}

func (cfg *TemporalClientConfig) Check() {
	if cfg.HostPort == "" {
		log.Fatal("[fatal] Temporal client HostPort is not set")
	}
	if cfg.Namespace == "" {
		log.Fatal("[fatal] Temporal client Namespace is not set")
	}
}

func NewClient(cfg TemporalClientConfig) (client.Client, error) {
	cfg.Check()

	options := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	}

	return client.NewLazyClient(options)
}
