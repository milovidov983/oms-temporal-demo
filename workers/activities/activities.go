package activities

import (
	"context"
	"log"
)

type ActivitiesConfig struct {
	OmsCoreHostPort string
}

func (cfg *ActivitiesConfig) Check() {
	if cfg.OmsCoreHostPort == "" {
		log.Fatal("[fatal] Activities OmsCoreHostPort is not set")
	}
}

type Activities struct {
	OmsCoreHost string
}

func NewActivities(cfg *ActivitiesConfig) *Activities {
	cfg.Check()

	return &Activities{
		OmsCoreHost: cfg.OmsCoreHostPort,
	}
}

type AssemblyApplicationInput struct {
	OrderID string
}

func (a *Activities) CreateAssemblyApplication(ctx context.Context, input *AssemblyApplicationInput) error {

	return nil
}
