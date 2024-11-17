package main

import (
	"context"
	"log"
	"os"

	"github.com/milovidov983/oms-temporal-demo/temporal-adapter/consumer"
	"github.com/milovidov983/oms-temporal-demo/temporal-adapter/handler"
	"github.com/spf13/viper"
)

func main() {
	log.SetPrefix("[temporal-adapter]:")

	loadConfig()

	handlerConfig := handler.OrderHandlerConfig{
		TemporalHost: viper.GetString("temporal.hostPort"),
		Namespace:    viper.GetString("temporal.oms"),
	}
	handlerConfig.Check()

	orderHandler, err := handler.NewOrderHandler(handlerConfig)
	if err != nil {
		log.Fatalf("[fatal] Error creating order handler: %v", err)
	}
	cosumerConfig := consumer.ConsumerConfig{
		Brokers: []string{viper.GetString("kafka.brokers")},
		GroupID: viper.GetString("kafka.consumerGroup"),
		Topic:   "oms.oms-core.orders.v1",
		Handler: orderHandler,
	}
	cosumerConfig.Check()

	consumer, err := consumer.NewKafkaConsumer(cosumerConfig)
	if err != nil {
		log.Fatalf("Error creating consumer: %v", err)
	}

	ctx := context.Background()
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Error running consumer: %v", err)
	}
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
