package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/worker"

	"github.com/milovidov983/oms-temporal-demo/workers/activities"
	"github.com/milovidov983/oms-temporal-demo/workers/temporal"
	"github.com/milovidov983/oms-temporal-demo/workers/workflows"
	"github.com/spf13/viper"
)

func main() {
	log.SetPrefix("[temporal-worker]:")

	loadConfig()

	c, err := temporal.NewClient()

	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "order-processing-main", worker.Options{})

	w.RegisterWorkflow(workflows.ProcessOrder)

	w.RegisterActivity(&activities.Activities{
		OmsCoreHost: "localhost:8889",
	})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("worker exited: %v", err)
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
