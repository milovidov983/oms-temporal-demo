package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/milovidov983/oms-temporal-demo/shared/models"
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

type Input struct {
	OrderID string
}

func (a *Activities) CreateAssemblyApplication(ctx context.Context, input *Input) (string, error) {
	url := "http://" + a.OmsCoreHost + "/api/assembly"

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

func (a *Activities) GetOrderTypes(ctx context.Context, input *Input) ([]models.OrderType, error) {

	// Тут мы ходим в oms-core за свойствами заказа, условно, надо его доставлять собирать и так далее.
	// Ожнако в целевой продовой версии можно передавать все эти свойства в workflow в качестве
	// входящих параметров, а не делать запросы к oms-core.
	// Если один из признаков удаляется, например клиент решил приехать сам за заказом. То это свойство
	// можно поменять послав соответсвующий сигнал в wkrkflow

	// mock: для целей демонстраии, все заказы со сборкой и доставкой
	orderTypes := []models.OrderType{
		models.OrderTypeDelivery,
		models.OrderTypeAssembly,
	}

	return orderTypes, nil
}
