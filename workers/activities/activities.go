package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (a *Activities) CreateAssemblyApplication(ctx context.Context, input *AssemblyApplicationInput) (string, error) {
	url := a.OmsCoreHost + "/assembly"

	request := struct {
		OrderID string `json:"order_id"`
	}{
		OrderID: input.OrderID,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var responseBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", err
	}

	applicationID, ok := responseBody["application_id"]
	if !ok {
		return "", fmt.Errorf("missing application_id in response")
	}

	return applicationID, nil
}
