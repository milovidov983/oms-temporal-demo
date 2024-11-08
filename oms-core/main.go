package main

import (
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/milovidov983/oms-temporal-demo/oms-core/handler"
	"github.com/milovidov983/oms-temporal-demo/oms-core/repository"
	"github.com/milovidov983/oms-temporal-demo/oms-core/service"
	"github.com/spf13/viper"
)

func main() {
	log.SetPrefix("[oms-core]:")
	loadConfig()

	dbConnectionString := viper.GetString("database.connectionString")
	repo, err := repository.NewOrderRepository(dbConnectionString)
	if err != nil {
		log.Fatalf("[fatal] Error creating order repository: %v", err)
	}
	log.Printf("[info] Order repository created")

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	brokerAddresses := viper.GetStringSlice("kafka.brokers")
	log.Printf("[info] Kafka brokers: %v", brokerAddresses)
	producer, err := sarama.NewSyncProducer(brokerAddresses, config)
	if err != nil {
		log.Fatalf("[fatal] Error creating Kafka producer: %v", err)
	}
	defer producer.Close()

	orderTopic := viper.GetString("kafka.topic")
	orderService := service.NewOrderService(repo, producer, orderTopic)

	log.Printf("[info] Order service created with topic: %s", orderTopic)

	orderHandler := handler.NewOrderHandler(orderService)
	http.HandleFunc("/api/orders", orderHandler.CreateOrder)
	http.HandleFunc("/api/orders/status", orderHandler.GetStatus)

	port := viper.GetString("server.address")
	log.Printf("[info] Starting server on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
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
