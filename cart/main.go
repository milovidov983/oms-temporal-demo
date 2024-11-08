package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
)

type TextResponse struct {
	Text string `json:"text"`
}

type OrderRequest struct {
	Items []struct {
		ProductId string  `json:"product_id"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
	} `json:"items"`
}

type OrderResponse struct {
	OrderId string `json:"order_id"`
}

func main() {
	log.SetPrefix("[cart]:")
	log.Println("[info] Cart service starting..")

	loadConfig()

	http.HandleFunc("/api/cart/status", getStatus)
	http.HandleFunc("/api/cart", createOrder)

	host := viper.GetString("server.host")
	port := viper.GetInt("server.port")

	log.Printf("[info] Cart service started on %s:%d", host, port)

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}

func loadConfig() {
	env := os.Getenv("APP_ENV")
	configName := "config.dev.yml"
	if env == "production" {
		configName = "config.prod.yml"
	}

	log.Printf("[info] Loading configuration for environment: %s", env)

	viper.SetConfigName(configName)
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("[fatal] Config file not found: %v", err)
		} else {
			log.Fatalf("[fatal] Error reading config file: %v", err)
		}
	}

	viper.SetEnvPrefix("oms")
	viper.AutomaticEnv()

	log.Printf("[info] Configuration loaded successfully from: %s", viper.ConfigFileUsed())
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("[info] Request received %s", r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Query().Get("order_id")

	omsCoreHost := viper.GetString("external.services.oms-core.host")
	omsCorePort := viper.GetInt("external.services.oms-core.port")
	omsCoreAddress := fmt.Sprintf("%s:%d", omsCoreHost, omsCorePort)

	omsGetStatusUrl := fmt.Sprintf("http://%s/api/orders/status?order_id=%s", omsCoreAddress, orderID)
	log.Printf("[info] make call to %s", omsGetStatusUrl)
	resp, err := http.Get(omsGetStatusUrl)
	if err != nil {
		log.Printf("[error] Error making request to OMS Core: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("[error] Unexpected status code from OMS Core: %d", resp.StatusCode)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[error] Error reading response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var orderStatus struct {
		Status string `json:"status"`
	}
	err = json.Unmarshal(body, &orderStatus)
	if err != nil {
		log.Printf("[error] Error unmarshalling response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := TextResponse{
		Text: fmt.Sprintf("Order status: %s", orderStatus.Status),
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("[info] Request processed successfully")
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	log.Printf("[info] Request received %s", r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[error] Error reading request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var orderRequest OrderRequest
	err = json.Unmarshal(body, &orderRequest)
	if err != nil {
		log.Printf("[error] Error unmarshalling request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	omsCoreRequest := struct {
		CustomerId string `json:"customer_id"`
		Items      []struct {
			ProductId string  `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
		CreatedAt string `json:"created_at"`
	}{
		CustomerId: "customer456",
		Items:      orderRequest.Items,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	omsCoreHost := viper.GetString("external.services.oms-core.host")
	omsCorePort := viper.GetInt("external.services.oms-core.port")
	omsCoreAddress := fmt.Sprintf("%s:%d", omsCoreHost, omsCorePort)

	omsBodyRequest, err := json.Marshal(omsCoreRequest)
	if err != nil {
		log.Printf("[error] Error marshalling request body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	omsCreateOrderUrl := fmt.Sprintf("http://%s/api/orders", omsCoreAddress)
	log.Printf("[debug] Make call to %s", omsCreateOrderUrl)
	resp, err := http.Post(omsCreateOrderUrl, "application/json", bytes.NewBuffer(omsBodyRequest))
	if err != nil {
		log.Printf("[error] Error making request to OMG Core: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[error] Error reading response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var orderResponse OrderResponse
	err = json.Unmarshal(respBody, &orderResponse)
	if err != nil {
		log.Printf("[error] Error unmarshalling response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(orderResponse)
	if err != nil {
		log.Printf("[error] Error encoding response body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("[info] Request processed successfully")
}
