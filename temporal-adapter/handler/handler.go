package handler

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
)

type HandlerConfig struct {
	TemporalHost string
	Namespace    string
}

func (cfg *HandlerConfig) Check() {
	if cfg.TemporalHost == "" {
		log.Fatal("[fatal] Temporal host is not set")
	}
	if cfg.Namespace == "" {
		log.Fatal("[fatal] Temporal Namespace is not set")
	}
}

type Handler struct {
	logger   *log.Logger
	temporal client.Client
}

func newTemporalClient(options client.Options) (client.Client, error) {
	if options.HostPort == "" {
		options.HostPort = os.Getenv("TEMPORAL_GRPC_ENDPOINT")
	}

	return client.NewLazyClient(options)
}

func NewHandler(cfg HandlerConfig) (*Handler, error) {
	client, err := newTemporalClient(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		return nil, err
	}
	return &Handler{
		logger:   log.New(os.Stdout, "[order-handler]", log.LstdFlags),
		temporal: client,
	}, nil
}
