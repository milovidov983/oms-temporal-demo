package main

import (
	"database/sql"
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
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("failed to connect to database: %w", err)
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	brokerAddresses := viper.GetStringSlice("kafka.brokers")
	log.Printf("[info] Kafka brokers: %v", brokerAddresses)
	producer, err := sarama.NewSyncProducer(brokerAddresses, config)
	if err != nil {
		log.Fatalf("[fatal] Error creating Kafka producer: %v", err)
	}
	defer producer.Close()

	// Order
	orderRepo, err := repository.NewOrderRepository(db)
	if err != nil {
		log.Fatalf("[fatal] Error creating order repository: %v", err)
	}
	log.Printf("[info] Order repository created")
	orderTopic := viper.GetString("kafka.topics.order")
	orderService := service.NewOrderService(orderRepo, producer, orderTopic)

	log.Printf("[info] Order service created with topic: %s", orderTopic)

	orderHandler := handler.NewOrderHandler(orderService)
	http.HandleFunc("/api/orders", orderHandler.CreateOrder)
	http.HandleFunc("/api/orders/status", orderHandler.GetStatus)

	// Assembly
	assRepo, err := repository.NewAssemblyApplicationRepository(db)
	if err != nil {
		log.Fatalf("[fatal] Error creating order repository: %v", err)
	}
	log.Printf("[info] Assembly appliation repository created")

	assemblyApplicationTopic := viper.GetString("kafka.topics.assemblyApplication")
	assemblyApplicationService := service.NewAssemblyApplicationService(assRepo, producer, assemblyApplicationTopic)
	assemblyHandler := handler.NewAssemblyApplicationHandler(assemblyApplicationService)
	http.HandleFunc("/api/assembly", assemblyHandler.CreateApplication)
	http.HandleFunc("/api/assembly/complete", assemblyHandler.CompleteApplication)
	http.HandleFunc("/api/assembly/cancel", assemblyHandler.CompleteApplication)

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
